package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vijaynallagatla/vjvector/internal/api"
	"github.com/vijaynallagatla/vjvector/internal/server"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("VJVECTOR_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Create server instance
	srv, err := server.NewServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Create API handlers
	handlers := api.NewHandlers()

	// Register API routes
	handlers.RegisterRoutes(srv.Echo())

	// Start the server
	if err := srv.Start(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown error: %v\n", err)
	}
}
