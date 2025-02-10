package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"time"
	"zeropii/db"
	"zeropii/models"
	"zeropii/utils"
)

var client *mongo.Client
var customerCollection *mongo.Collection

//var encryptionKey = os.Getenv("ENCRYPTION_KEY")

func initMongoDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err2 error
	client, err2 = mongo.Connect(ctx, clientOptions)
	if err2 != nil {
		log.Fatal().Err(err2).Msg("Failed to connect to MongoDB: %v")
	}
	customerCollection = client.Database("onboarding").Collection("customer")

}

func main() {
	// Initialize logger
	initLogger()

	// Load environment and initialize mongo
	loadEnv()
	db.InitMongoDB()

	router := gin.Default()

	//Apply request logger middleware
	router.Use(requestLoggerMiddleware())

	// Customer end points
	api := router.Group("/api/v1/onboarding")
	{
		api.POST("/customers", createCustomer)
		api.GET("/customers/:id", getCustomer)
		//api.GET("/api/v1/customers/partner/:partnerID", "")
	}

	// Start the server
	port := os.Getenv("PORT")
	err := router.Run(":" + port)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "start_server").
			Msg("Failed to start server")
		return
	}
}

// Create a new customer
func createCustomer(c *gin.Context) {
	var customer models.Customer

	// Bind the incoming JSON to the customer model
	if err := c.BindJSON(&customer); err != nil {
		log.Error().
			Err(err).
			Str("operation", "create_customer").
			Msg("Failed to bind customer JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the customer creation attempt (mask PII like SSN)
	log.Info().
		Str("operation", "create_customer").
		Str("name", customer.FullName).
		Str("email", customer.Email). // Don't log PII data
		Msg("Creating new customer")

	// Encrypt PII before storing it
	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	// Encrypt PII data
	if err := utils.EncryptStructPII(&customer, encryptionKey); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to encrypt customer PII")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt customer PII"})
		return
	}

	// Check if CustomerID is empty, and assign a new UUID if so
	if customer.ID == "" {
		customer.ID = uuid.New().String()
	}

	// Insert the customer into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := customerCollection.InsertOne(ctx, customer)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "insert_customer").
			Msg("Failed to insert customer into MongoDB")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	// Log success
	log.Info().
		Str("operation", "insert_customer").
		Str("customer_id", result.InsertedID.(string)).
		Msg("Customer successfully created")

	// Return response
	c.JSON(http.StatusCreated, gin.H{"customer_id": result.InsertedID})
}

// Get a customer by ID
func getCustomer(c *gin.Context) {
	id := c.Param("id")

	// Log the request to get a customer
	log.Info().
		Str("operation", "get_customer").
		Str("customer_id", id).
		Msg("Fetching customer details")

	var customer models.Customer

	userRole := c.GetHeader("x-viewer-role")
	if userRole == "" {
		log.Error().Msg("x-viewer-role header not provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Viewer role header is required"})
		return
	}

	// Find the customer by ID
	err := customerCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&customer)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	// Decrypt PII data
	if err := utils.DecryptStructPII(&customer, encryptionKey); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to decrypt customer PII")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt customer PII"})
		return
	}

	// Mask sensitive information before logging
	log.Info().
		Str("operation", "get_customer").
		Str("name", customer.FullName).
		Str("email", customer.Email).
		Msg("Customer retrieved successfully")

	// Sanitize PII fields based on the user's role
	utils.SanitizeCustomerData(&customer, userRole)

	c.JSON(http.StatusOK, customer)
	logResponse(c, http.StatusOK, customer)
}

func getCustomerByPartnerID(c *gin.Context) {
	partnerID := c.Param("partner_id")

	if partnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Partner ID is required"})
		return
	}

	// Log the request to get a customer
	log.Info().
		Str("operation", "get_customer_by_partner_id").
		Str("partner_id", partnerID).
		Msg("Fetching customer details by partner")

	var customer models.Customer

	userRole := c.GetHeader("x-viewer-role")
	if userRole == "" {
		log.Error().Msg("x-viewer-role header not provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Viewer role header is required"})
		return
	}

	// Find by partner id
	cur, err := customerCollection.Find(context.Background(), bson.M{"partner_id": partnerID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}(cur, context.Background())

	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	// Decrypt PII data
	if err := utils.DecryptStructPII(&customer, encryptionKey); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to decrypt customer PII")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt customer PII"})
		return
	}

	// Mask sensitive information before logging
	log.Info().
		Str("operation", "get_customer").
		Str("name", customer.FullName).
		Str("email", customer.Email).
		Msg("Customer retrieved successfully")

	// Sanitize PII fields based on the user's role
	utils.SanitizeCustomerData(&customer, userRole)

	c.JSON(http.StatusOK, customer)
	logResponse(c, http.StatusOK, customer)
}

func loadEnv() {
	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env fil")
	}
}

func initLogger() {
	// Output logs to both console and file (server will be /var/log/customer_api.log)
	logFile, err := os.OpenFile("customer_api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}

	// Set Zerolog output to both console and log file
	log.Logger = zerolog.New(zerolog.MultiLevelWriter(logFile, zerolog.ConsoleWriter{Out: os.Stdout})).
		With().
		Timestamp().
		Logger()

	// Optionally, set a human-readable time format for debugging
	zerolog.TimeFieldFormat = time.RFC3339
}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process the request
		c.Next()

		// Log the request details
		duration := time.Since(startTime)
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Msg("incoming request")
	}
}

func logResponse(c *gin.Context, statusCode int, response interface{}) {
	log.Info().
		Str("operation", "response").
		Int("status_code", statusCode).
		Interface("response_body", response).
		Msg("Response sent to client")
}
