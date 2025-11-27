package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/homecraft/backend/pkg/handlers"
	"github.com/homecraft/backend/pkg/k8s"
)

func main() {
	// Create Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create server handler
	serverHandler := handlers.NewServerHandler(k8sClient)

	// Set Gin mode from environment
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", serverHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Minecraft server endpoints
		v1.POST("/servers", serverHandler.CreateServer)
		v1.GET("/servers", serverHandler.ListServers)
		v1.GET("/servers/:name", serverHandler.GetServer)
		v1.DELETE("/servers/:name", serverHandler.DeleteServer)

		// Cluster resource endpoints
		v1.GET("/cluster/resources", serverHandler.GetClusterResources)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HomeCraft API server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
