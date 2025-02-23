package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"company-microservice/handler"
	"company-microservice/kafkaproducer"
	"company-microservice/middleware"
)

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

	a, err := middleware.NewJwtAuth(db)
	if err != nil {
		log.Fatalf("Failed to create JWT Auth %v", err)
		return
	}

	// Start the server
	r := gin.Default()

	// Public routes
	r.POST("/token", a.GenerateToken)
	r.GET("/companies/:id", h.GetCompany)

	// Protected routes
	protected := r.Group("/")
	protected.Use(a.AuthMiddleware())
	{
		protected.POST("/companies", h.CreateCompany)
		protected.PATCH("/companies/:id", h.UpdateCompany)
		protected.DELETE("/companies/:id", h.DeleteCompany)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server running returned error: %v\n", err)
	}
}
