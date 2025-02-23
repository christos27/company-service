package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"company-microservice/handler"
	"company-microservice/kafkaproducer"
)

// JWT middleware
func AuthMiddleware() gin.HandlerFunc {
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

		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Generate JWT token (for testing)
func GenerateToken(c *gin.Context) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

var jwtSecret = []byte("your-secret-key")

var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbName = os.Getenv("DB_NAME")
var kafkaBroker = os.Getenv("KAFKA_BROKER")
var kafkaTopic = os.Getenv("KAFKA_TOPIC")

func main() {
	var err error
	var kafkaProducer *kafkaproducer.Producer
	var db *pgxpool.Pool

	// initialize database connection
	var connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Initialize Kafka producer
	kafkaProducer, err = kafkaproducer.NewProducer(kafkaBroker, kafkaTopic)
	if err != nil {
		log.Fatalf("Failed to start Kafka producer: %v", err)
		return
	}
	defer kafkaProducer.Close()

	// Create the CRUD handlers
	h, err := handler.NewHandler(db, kafkaProducer)
	if err != nil {
		log.Fatalf("Failed to create handlers %v", err)
		return
	}

	// Start the server
	r := gin.Default()

	// Public routes
	r.POST("/token", GenerateToken)
	r.GET("/companies/:id", h.GetCompany)

	// Protected routes
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/companies", h.CreateCompany)
		protected.PATCH("/companies/:id", h.UpdateCompany)
		protected.DELETE("/companies/:id", h.DeleteCompany)
	}

	r.Run(":8080")
}
