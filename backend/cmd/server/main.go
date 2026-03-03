package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	indexallv1connect "github.com/construct/indexall/internal/gen/pb/indexall/v1/indexallv1connect"
	"github.com/construct/indexall/internal/db"
	"github.com/construct/indexall/internal/service"
)

func main() {
	// Configuration
	dbPath := "indexall.db"
	port := 8080

	// Parse flags or environment variables
	if path := os.Getenv("DB_PATH"); path != "" {
		dbPath = path
	}
	if portStr := os.Getenv("PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	fmt.Printf("Starting IndexAll server...\n")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Port: %d\n", port)

	// Initialize database
	database, err := db.InitDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Get queries instance
	queries := db.GetQueries(database)

	// Create services
	tagService := service.NewTagService(database, queries)
	resourceService := service.NewResourceService(database, queries)

	// Create HTTP mux and register services
	mux := http.NewServeMux()

	// Register ConnectRPC services
	mux.Handle(indexallv1connect.NewTagServiceHandler(tagService))
	mux.Handle(indexallv1connect.NewResourceServiceHandler(resourceService))

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        logMiddleware(mux),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("✓ Server listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	fmt.Println("\nShutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Server stopped gracefully")
}

// logMiddleware logs HTTP requests
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
