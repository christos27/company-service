package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

	company.ID = uuid.New().String()

	query := `INSERT INTO companies (id, name, description, amount_of_employees, registered, type)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	_, err := db.Exec(context.Background(), query, company.ID, company.Name, company.Description, company.AmountOfEmployees, company.Registered, company.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert company"})
		return
	}

	c.JSON(http.StatusCreated, company)
}

// Get company by ID
func GetCompany(c *gin.Context) {
	id := c.Param("id")

	var company Company
	query := `SELECT id, name, description, amount_of_employees, registered, type FROM companies WHERE id=$1`
	err := db.QueryRow(context.Background(), query, id).Scan(&company.ID, &company.Name, &company.Description, &company.AmountOfEmployees, &company.Registered, &company.Type)
	if err != nil {
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

	var updateData UpdateCompanyRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start building the SQL query dynamically
	query := "UPDATE companies SET"
	var args []interface{}
	argIndex := 1

	if updateData.Name != nil {
		query += fmt.Sprintf(" name = $%d,", argIndex)
		args = append(args, *updateData.Name)
		argIndex++
	}
	if updateData.Description != nil {
		query += fmt.Sprintf(" description = $%d,", argIndex)
		args = append(args, *updateData.Description)
		argIndex++
	}
	if updateData.AmountOfEmployees != nil {
		query += fmt.Sprintf(" amount_of_employees = $%d,", argIndex)
		args = append(args, *updateData.AmountOfEmployees)
		argIndex++
	}
	if updateData.Registered != nil {
		query += fmt.Sprintf(" registered = $%d,", argIndex)
		args = append(args, *updateData.Registered)
		argIndex++
	}
	if updateData.Type != nil {
		query += fmt.Sprintf(" type = $%d,", argIndex)
		args = append(args, *updateData.Type)
		argIndex++
	}

	// If no fields were provided, return an error
	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// Remove the last comma and add WHERE clause
	query = query[:len(query)-1] + fmt.Sprintf(" WHERE id = $%d", argIndex) + " RETURNING  id, name, description, amount_of_employees, registered, type"
	args = append(args, id)

	// Execute the query
	var company Company
	err := db.QueryRow(context.Background(), query, args...).Scan(&company.ID, &company.Name, &company.Description, &company.AmountOfEmployees, &company.Registered, &company.Type)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}

	c.JSON(http.StatusOK, company)
}

// Delete company
func DeleteCompany(c *gin.Context) {
	id := c.Param("id")

	query := `DELETE FROM companies WHERE id=$1`
	_, err := db.Exec(context.Background(), query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}

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

var db *pgxpool.Pool

func main() {
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://postgres:password@172.17.0.2:5432/companies")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		return
	}
	defer db.Close()

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
