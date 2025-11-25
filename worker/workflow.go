package main

import (
	"fmt"
	"time"

	"temporal-cloud/shared"

	"go.temporal.io/sdk/workflow"
)

// ScheduleWorkflowInput defines the payload for the schedule-triggered workflow
type ScheduleWorkflowInput struct {
	Name string // human-readable name parameter
}

// DelayWorkflowInput defines the payload for the delay workflow
type DelayWorkflowInput struct {
	DelayMinutes int    // number of minutes to delay before waiting for a signal
	Name         string // human-readable name parameter
}

// ScheduleWorkflow is executed by the Temporal schedule trigger
// Once triggered it runs to completion immediately (cron-style automation)
func ScheduleWorkflow(ctx workflow.Context, input ScheduleWorkflowInput) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ScheduleWorkflow started", "input", input)

	name := input.Name
	if name == "" {
		name = shared.DefaultWorkflowName
	}

	// Run immediately and finish—typical cron-task behavior, no signal needed
	result := fmt.Sprintf("Hello, %s from Temporal Worker on AWS!", name)
	logger.Info("Completing workflow", "result", result)

	return result, nil
}

// DelayWorkflow applies a delay before waiting for a signal and then completing
func DelayWorkflow(ctx workflow.Context, input DelayWorkflowInput) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("DelayWorkflow started", "input", input)

	delayMinutes := input.DelayMinutes
	if delayMinutes <= 0 {
		delayMinutes = 1 // default to 1 minute
	}

	name := input.Name
	if name == "" {
		name = shared.DefaultWorkflowName
	}

	logger.Info("Waiting for delay", "minutes", delayMinutes)

	// 1. Sleep for the requested number of minutes
	delayDuration := time.Duration(delayMinutes) * time.Minute
	workflow.Sleep(ctx, delayDuration)

	logger.Info("Delay completed, now waiting for signal", "name", name)

	// 2. After the delay, begin waiting for the signal
	signalChan := workflow.GetSignalChannel(ctx, "update-name")
	logger.Info("Waiting for signal 'update-name' to update name...")

	var newName string
	// Block here until the client sends the signal
	more := signalChan.Receive(ctx, &newName)
	if !more {
		logger.Info("Signal channel closed")
	} else {
		logger.Info("Received signal", "newName", newName)
	}

	if newName != "" {
		name = newName
	}

	// Once the signal is received, return the result
	result := fmt.Sprintf("Hello, %s from Temporal Worker on AWS!", name)
	logger.Info("Completing workflow", "result", result)

	return result, nil
}

func HelloWorkflow(ctx workflow.Context, input interface{}) (string, error) {
	logger := workflow.GetLogger(ctx)

	// Support both legacy string arguments and the newer WorkflowInput struct
	var name string
	var autoStart bool

	switch v := input.(type) {
	case string:
		// Legacy path: raw string argument
		name = v
		autoStart = false
	case shared.WorkflowInput:
		// Newer path: structured input
		name = v.Name
		autoStart = v.AutoStart
	default:
		// Fallback default
		name = shared.DefaultWorkflowName
		autoStart = false
	}

	logger.Info("HelloWorkflow started", "name", name, "autoStart", autoStart)

	// When autoStart is true (schedule scenario), run immediately without waiting
	if autoStart {
		logger.Info("Auto-start mode: executing immediately without waiting for signal")
		result := fmt.Sprintf("Hello, %s from Temporal Worker on AWS!", name)
		logger.Info("Completing workflow", "result", result)
		return result, nil
	}

	// Manual start flows wait for signals
	signalChan := workflow.GetSignalChannel(ctx, "update-name")
	logger.Info("Waiting for signal 'update-name' to update name...")

	var newName string
	// Block here until the client sends the signal
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

	// After the signal, finish and return the result
	result := fmt.Sprintf("Hello, %s from Temporal Worker on AWS!", name)
	logger.Info("Completing workflow", "result", result)

	return result, nil
}
