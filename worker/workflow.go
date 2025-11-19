package main

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"temporal-cloud/shared"
)

func HelloWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorkflow started", "name", name)

	signalChan := workflow.GetSignalChannel(ctx, "update-name")

	logger.Info("Waiting for signal 'update-name' to update name...")

	var newName string
	// 这里会挂住，直到 client 发送 Signal
	// Receive will block until signal is received
	more := signalChan.Receive(ctx, &newName)
	if !more {
		logger.Info("Signal channel closed")
	} else {
		logger.Info("Received signal", "newName", newName)
	}

	if newName != "" {
		name = newName
	}

	// 2. 收到 signal 后再返回结果
	result := fmt.Sprintf("Hello, %s from Temporal Worker on AWS!", name)
	logger.Info("Completing workflow", "result", result)

	return result, nil
}
