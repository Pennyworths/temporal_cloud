package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
)

func main() {
	var envPath string
	if _, err := os.Stat(".env"); err == nil {
		envPath = ".env"
	} else {
		envPath = filepath.Join("..", ".env")
	}

	if err := godotenv.Load(envPath); err != nil {
		log.Println("No .env file found, using environment variables only")
	}

	addr := os.Getenv("TEMPORAL_ADDRESS")
	ns := os.Getenv("TEMPORAL_NAMESPACE")
	apiKey := os.Getenv("TEMPORAL_API_KEY")
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")

	if addr == "" || ns == "" || apiKey == "" || taskQueue == "" {
		log.Fatal("Missing required Temporal configuration")
	}

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

	log.Println("✅ Client connected to Temporal Cloud")

	ctx := context.Background()
	workflowID := fmt.Sprintf("hello-workflow-%d", time.Now().Unix())

	wo := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}

	name := "xxx-xxx"
	we, err := c.ExecuteWorkflow(ctx, wo, "HelloWorkflow", name)
	if err != nil {
		log.Fatalf("Unable to start workflow: %v", err)
	}

	log.Printf("🚀 Started workflow. ID=%s RunID=%s\n", we.GetID(), we.GetRunID())
	var result string
	if err := we.Get(ctx, &result); err != nil {
		log.Fatalf("Unable to get workflow result: %v", err)
	}
	log.Printf("🎉 Workflow completed. Result: %s\n", result)
}
