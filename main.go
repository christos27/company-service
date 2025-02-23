package main

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte("your-secret-key")

// Company struct
type Company struct {
	ID                string `json:"id"`
	Name              string `json:"name" binding:"required,max=15"`
	Description       string `json:"description,omitempty" binding:"max=3000"`
	AmountOfEmployees int    `json:"amountOfEmployees" binding:"required"`
	Registered        bool   `json:"registered" binding:"required"`
	Type              string `json:"type" binding:"required,oneof=Corporations NonProfit Cooperative Sole_Proprietorship"`
}

// In-memory storage
var companies = make(map[string]Company)
var mu sync.Mutex

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

// Create company
func CreateCompany(c *gin.Context) {
	var company Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	company.ID = uuid.New().String()
	if _, exists := companies[company.Name]; exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company name already exists"})
		return
	}
	companies[company.ID] = company

	c.JSON(http.StatusCreated, company)
}

// Get company by ID
func GetCompany(c *gin.Context) {
	id := c.Param("id")

	mu.Lock()
	defer mu.Unlock()

	company, exists := companies[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	c.JSON(http.StatusOK, company)
}

// Use pointer fields to allow partial updates
type UpdateCompanyRequest struct {
	Name              *string `json:"name" binding:"omitempty,max=15"`
	Description       *string `json:"description,omitempty" binding:"omitempty,max=3000"`
	AmountOfEmployees *int    `json:"amountOfEmployees,omitempty"`
	Registered        *bool   `json:"registered,omitempty"`
	Type              *string `json:"type,omitempty" binding:"omitempty,oneof=Corporations NonProfit Cooperative Sole_Proprietorship"`
}

// Update company
func UpdateCompany(c *gin.Context) {
	id := c.Param("id")

	mu.Lock()
	defer mu.Unlock()

	company, exists := companies[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	var updateData UpdateCompanyRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
	if updateData.Name != nil {
		company.Name = *updateData.Name
	}
	if updateData.Description != nil {
		company.Description = *updateData.Description
	}
	if updateData.AmountOfEmployees != nil {
		company.AmountOfEmployees = *updateData.AmountOfEmployees
	}
	if updateData.Registered != nil {
		company.Registered = *updateData.Registered
	}
	if updateData.Type != nil {
		company.Type = *updateData.Type
	}

	companies[id] = company
	c.JSON(http.StatusOK, company)
}

// Delete company
func DeleteCompany(c *gin.Context) {
	id := c.Param("id")

	mu.Lock()
	defer mu.Unlock()

	if _, exists := companies[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	delete(companies, id)
	c.JSON(http.StatusNoContent, nil)
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

func main() {
	r := gin.Default()

	// Public routes
	r.POST("/token", GenerateToken)
	r.GET("/companies/:id", GetCompany)

	// Protected routes
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/companies", CreateCompany)
		protected.PATCH("/companies/:id", UpdateCompany)
		protected.DELETE("/companies/:id", DeleteCompany)
	}

	r.Run(":8080")
}
