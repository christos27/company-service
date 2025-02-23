package handler

import (
	"company-microservice/kafkaproducer"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db            *pgxpool.Pool
	kafkaProducer *kafkaproducer.Producer
}

func NewHandler(db *pgxpool.Pool, kafkaProducer *kafkaproducer.Producer) (*Handler, error) {
	if db == nil {
		return nil, errors.New("Database connection was not provided")
	}
	if kafkaProducer == nil {
		return nil, errors.New("Kafka connection was not provided")
	}
	return &Handler{
		db:            db,
		kafkaProducer: kafkaProducer,
	}, nil
}

// Company struct
type Company struct {
	ID                string `json:"id"`
	Name              string `json:"name" binding:"required,max=15"`
	Description       string `json:"description,omitempty" binding:"max=3000"`
	AmountOfEmployees int    `json:"amountOfEmployees" binding:"required"`
	Registered        *bool  `json:"registered" binding:"required"`
	Type              string `json:"type" binding:"required,oneof=Corporations NonProfit Cooperative Sole_Proprietorship"`
}

// Create company
func (h *Handler) CreateCompany(c *gin.Context) {
	var company Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company.ID = uuid.New().String()

	registered := false // Default value
	if company.Registered != nil {
		registered = *company.Registered
	}

	query := `INSERT INTO companies (id, name, description, amount_of_employees, registered, type)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	_, err := h.db.Exec(context.Background(), query, company.ID, company.Name, company.Description, company.AmountOfEmployees, registered, company.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert company"})
		return
	}

	// Produce Kafka event
	event := fmt.Sprintf(`{"event": "company_created", "id": "%s", "name": "%s"}`, company.ID, company.Name)
	err = h.kafkaProducer.ProduceMessage(company.ID, event)
	if err != nil {
		log.Printf("Failed to send Kafka event: %v", err)
	}

	c.JSON(http.StatusCreated, company)
}

// Get company by ID
func (h *Handler) GetCompany(c *gin.Context) {
	id := c.Param("id")

	var company Company
	query := `SELECT id, name, description, amount_of_employees, registered, type FROM companies WHERE id=$1`
	err := h.db.QueryRow(context.Background(), query, id).Scan(&company.ID, &company.Name, &company.Description, &company.AmountOfEmployees, &company.Registered, &company.Type)
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
func (h *Handler) UpdateCompany(c *gin.Context) {
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
	err := h.db.QueryRow(context.Background(), query, args...).Scan(&company.ID, &company.Name, &company.Description, &company.AmountOfEmployees, &company.Registered, &company.Type)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}

	// Produce Kafka event
	event := fmt.Sprintf(`{"event": "company_updated", "id": "%s", "name": "%s"}`, company.ID, company.Name)
	err = h.kafkaProducer.ProduceMessage(id, event)
	if err != nil {
		log.Printf("Failed to send Kafka event: %v", err)
	}

	c.JSON(http.StatusOK, company)
}

// Delete company
func (h *Handler) DeleteCompany(c *gin.Context) {
	id := c.Param("id")

	var company Company
	query := `DELETE FROM companies WHERE id=$1 RETURNING id, name;`
	err := h.db.QueryRow(context.Background(), query, id).Scan(&company.ID, &company.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}

	// Produce Kafka event
	event := fmt.Sprintf(`{"event": "company_deleted", "id": "%s", "name": "%s"}`, company.ID, company.Name)
	err = h.kafkaProducer.ProduceMessage(id, event)
	if err != nil {
		log.Printf("Failed to send Kafka event: %v", err)
	}

	c.JSON(http.StatusNoContent, nil)
}
