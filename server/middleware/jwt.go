package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type JwtAuth struct {
	db *pgxpool.Pool
}

func NewJwtAuth(db *pgxpool.Pool) (*JwtAuth, error) {
	if db == nil {
		return nil, errors.New("Database connection was not provided")
	}
	return &JwtAuth{db: db}, nil
}

// Claims represents the JWT claims.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWT middleware
func (a *JwtAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// TODO We need to cache this
		var username string
		query := "SELECT username FROM admin_users WHERE username = $1"
		err = a.db.QueryRow(context.Background(), query, claims.Username).Scan(&username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

type TokenRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Generate JWT token (for testing)
func (a *JwtAuth) GenerateToken(c *gin.Context) {

	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check user in DB
	var userName string
	query := "SELECT username FROM admin_users WHERE username = $1 AND hashed_password = $2"
	err := a.db.QueryRow(context.Background(), query, req.Username, req.Password).
		Scan(&userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(30 * time.Minute)
	claims := &Claims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "some-issuer",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
