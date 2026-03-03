package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	indexallv1 "github.com/construct/indexall/internal/gen/pb/proto/indexall/v1"
	"github.com/construct/indexall/internal/db"
	"github.com/construct/indexall/internal/service"
)

func main() {
	// Configuration
	dbPath := "indexall.db"
	port := 50051 // Standard gRPC port

	// Parse flags or environment variables
	if path := os.Getenv("DB_PATH"); path != "" {
		dbPath = path
	}
	if portStr := os.Getenv("PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	fmt.Printf("Starting IndexAll gRPC server...\n")
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

	// Create gRPC server
	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	// Register services
	indexallv1.RegisterTagServiceServer(grpcServer, tagService)
	indexallv1.RegisterResourceServiceServer(grpcServer, resourceService)

	// Register reflection for debugging
	reflection.Register(grpcServer)

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen: %v\n", err)
		os.Exit(1)
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("✓ gRPC server listening on %s\n", listener.Addr())
		if err := grpcServer.Serve(listener); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	fmt.Println("\nShutting down server...")
	grpcServer.GracefulStop()

	fmt.Println("✓ Server stopped gracefully")
}
