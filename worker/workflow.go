package main

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func HelloWorkflow(ctx workflow.Context, name string) (string, error) {
	// Configure activity options (timeout etc.)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, SayHello, name).Get(ctx, &result)
	return result, err
}