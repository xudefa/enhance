// Package main demonstrates the gRPC starter usage.
//
// This example shows how to use the gRPC starter to:
// 1. Auto-configure gRPC server
// 2. Register services
// 3. Start the server
//
// Run:
//
//	go run main.go
//
// Test with grpcurl:
//
//	grpcurl -plaintext localhost:9090 list
package main

import (
	"context"
	"fmt"
	"net"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/xudefa/enhance/starter/grpc"
)

// Greeter is a simple gRPC service
type GreeterServer struct {
}

func (s *GreeterServer) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}

func main() {
	fmt.Println("=== gRPC Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("grpc-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	// Get the gRPC server from container
	grpcServer, err := core.GetByName[*grpc.Server](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get gRPC server: %v\n", err)
		return
	}

	// Register the Greeter service
	fmt.Println("--- Registering Services ---")
	greeterServer := &GreeterServer{}
	// In a real application, you would use generated code:
	// pb.RegisterGreeterServer(grpcServer, greeterServer)
	_ = greeterServer // Suppress unused variable warning

	// Enable reflection for debugging
	fmt.Println("Enabling gRPC reflection...")
	reflection.Register(grpcServer)

	// Start the server
	fmt.Println("\n--- Starting gRPC Server ---")
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}

	fmt.Println("gRPC server started on :9090")
	fmt.Println("Available services:")
	fmt.Println("  - grpc.reflection.v1alpha.ServerReflection")
	fmt.Println()
	fmt.Println("Test with grpcurl:")
	fmt.Println("  grpcurl -plaintext localhost:9090 list")
	fmt.Println("  grpcurl -plaintext localhost:9090 grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo")

	// Start serving in background
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			fmt.Printf("Failed to serve: %v\n", err)
		}
	}()

	// Wait for signal
	app.WaitForSignal()
}

// HelloRequest represents a hello request (placeholder for proto message)
type HelloRequest struct {
	Name string
}

// HelloReply represents a hello reply (placeholder for proto message)
type HelloReply struct {
	Message string
}
