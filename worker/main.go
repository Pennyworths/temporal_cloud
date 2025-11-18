package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Load environment variables (optional - ECS provides env vars directly)
	// Try to load .env file if it exists, but don't fail if it doesn't
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("Warning: Failed to load .env file: %v", err)
		}
	} else {
		log.Println("No .env file found, using environment variables from ECS")
	}

	addr := os.Getenv("TEMPORAL_ADDRESS")
	ns := os.Getenv("TEMPORAL_NAMESPACE")
	apiKey := os.Getenv("TEMPORAL_API_KEY")
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")

	if addr == "" || ns == "" || apiKey == "" || taskQueue == "" {
		log.Fatalf("Missing required Temporal configuration: addr=%v, ns=%v, apiKey=%v, taskQueue=%v",
			addr != "", ns != "", apiKey != "", taskQueue != "")
	}

	// Create Temporal client
	opts := client.Options{
		HostPort:  addr,
		Namespace: ns,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{},
		},
		Credentials: client.NewAPIKeyStaticCredentials(apiKey),
	}

	c, err := client.Dial(opts)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	log.Println("✅ Connected to Temporal Cloud!")

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(HelloWorkflow)
	w.RegisterActivity(SayHello)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Failed to run worker: %v", err)
	}
	log.Println("👋 Worker stopped")
}
