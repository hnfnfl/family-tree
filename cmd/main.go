package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hnfnfl/keluarga-tree/internal/config"
	"github.com/hnfnfl/keluarga-tree/internal/handlers"
	"github.com/hnfnfl/keluarga-tree/internal/middleware"
	"github.com/hnfnfl/keluarga-tree/internal/repository"
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
	defer driver.Close()

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}
	log.Println("✅ Connected to Neo4j")

	// Initialize repository
	personRepo := repository.NewPersonRepository(driver)
	familyRepo := repository.NewFamilyRepository(driver)
	userRepo := repository.NewUserRepository(driver)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWT.Secret)
	personHandler := handlers.NewPersonHandler(personRepo)
	familyHandler := handlers.NewFamilyHandler(familyRepo)

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "keluarga-tree-api",
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Public routes (no auth required)
		public := api.Group("")
		{
			public.GET("/families/:id/tree", familyHandler.GetFamilyTree)
			public.GET("/persons/search", personHandler.Search)
		}

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
			}

			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", authHandler.GetCurrentUser)
				users.PUT("/me", authHandler.UpdateProfile)
			}
		}
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("📥 Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped")
}
