package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hnfnfl/family-tree/internal/config"
	"github.com/hnfnfl/family-tree/internal/handlers"
	"github.com/hnfnfl/family-tree/internal/middleware"
	"github.com/hnfnfl/family-tree/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Neo4j driver
	driver, err := repository.NewNeo4jDriver(cfg.Neo4j)
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer closeCancel()
		driver.Close(closeCtx)
	}()

	// Verify connection
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer connectCancel()
	if err := driver.VerifyConnectivity(connectCtx); err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}
	log.Println("✅ Connected to Neo4j")

	// Initialize repositories
	personRepo := repository.NewPersonRepository(driver)
	familyRepo := repository.NewFamilyRepository(driver)
	userRepo := repository.NewUserRepository(driver)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWT.Secret)
	personHandler := handlers.NewPersonHandler(personRepo)
	familyHandler := handlers.NewFamilyHandler(familyRepo)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "keluarga-tree-api",
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Public routes (no auth required)
	api.GET("/families/:id/tree", familyHandler.GetFamilyTree)
	api.GET("/persons/search", personHandler.Search)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		// Person routes
		persons := protected.Group("/persons")
		{
			persons.GET("", personHandler.GetAll)
			persons.GET("/:id", personHandler.GetByID)
			persons.POST("", personHandler.Create)
			persons.PUT("/:id", personHandler.Update)
			persons.DELETE("/:id", personHandler.Delete)
			persons.POST("/:id/relate", personHandler.AddRelationship)
		}

		// Family routes
		families := protected.Group("/families")
		{
			families.GET("", familyHandler.GetAll)
			families.GET("/:id", familyHandler.GetByID)
			families.POST("", familyHandler.Create)
			families.PUT("/:id", familyHandler.Update)
			families.POST("/:id/members", familyHandler.AddMember)
		}

		// User routes
		users := protected.Group("/users")
		{
			users.GET("/me", authHandler.GetCurrentUser)
			users.PUT("/me", authHandler.UpdateProfile)
		}
	}

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Server starting on port %s (env: %s)", cfg.Server.Port, cfg.Server.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("📥 Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped")
}
