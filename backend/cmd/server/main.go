package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pdf-rag-system/backend/internal/api"
	"github.com/pdf-rag-system/backend/internal/client"
	"github.com/pdf-rag-system/backend/internal/repository"
	"github.com/pdf-rag-system/backend/internal/service"
	"github.com/pdf-rag-system/backend/pkg/config"
	"github.com/pdf-rag-system/backend/pkg/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize gRPC client
	docreaderClient, err := client.NewDocReaderClient(cfg.DocReader.Host, cfg.DocReader.Port)
	if err != nil {
		log.Fatalf("Failed to connect to docreader: %v", err)
	}
	defer docreaderClient.Close()

	// Initialize repositories
	documentRepo := repository.NewDocumentRepository(db)
	chunkRepo := repository.NewChunkRepository(db)

	// Initialize services
	documentService := service.NewDocumentService(documentRepo, chunkRepo, docreaderClient, cfg)
	chatService := service.NewChatService(chunkRepo, cfg)

	// Initialize handlers
	documentHandler := api.NewDocumentHandler(documentService)
	chatHandler := api.NewChatHandler(chatService)

	// Setup router
	router := gin.Default()

	// Set max multipart memory to 500MB (for file uploads)
	router.MaxMultipartMemory = 500 << 20 // 500 MB

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Document routes
		docs := v1.Group("/documents")
		{
			docs.POST("/upload", documentHandler.Upload)
			docs.GET("", documentHandler.List)
			docs.GET("/:id", documentHandler.Get)
			docs.GET("/:id/file", documentHandler.GetFile)
			docs.GET("/:id/page/:page/image", documentHandler.GetPageImage)
			docs.DELETE("/:id", documentHandler.Delete)
		}

		// Chat routes
		chat := v1.Group("/chat")
		{
			chat.POST("/query", chatHandler.Query)
		}
	}

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
