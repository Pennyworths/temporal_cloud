package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"temporal-cloud/shared"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
)

func main() {
	loadEnv()

	c := newTemporalClient()
	defer c.Close()

	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run client.go <start|signal|get|status|schedule|delay> [args...]")
	}

	command := os.Args[1]

	switch command {
	case "start":
		startWorkflow(c)
	case "signal":
		signalWorkflow(c)
	case "get":
		getWorkflowResult(c)
	case "status":
		getWorkflowStatus(c)
	case "schedule":
		handleScheduleCommand(c)
	case "delay":
		startDelayWorkflow(c)
	default:
		log.Fatalf("Unknown command %q. Use: start | signal | get | status | schedule | delay", command)
	}
}

func loadEnv() {
	var envPath string
	if _, err := os.Stat(".env"); err == nil {
		envPath = ".env"
	} else {
		envPath = filepath.Join("..", ".env")
	}

	if err := godotenv.Load(envPath); err != nil {
		log.Println("No .env file found, using environment variables only")
	}
}

func newTemporalClient() client.Client {
	addr := os.Getenv("TEMPORAL_ADDRESS")
	ns := os.Getenv("TEMPORAL_NAMESPACE")
	apiKey := os.Getenv("TEMPORAL_API_KEY")

	if addr == "" || ns == "" || apiKey == "" {
		log.Fatal("Missing required Temporal configuration")
	}

	c, err := client.Dial(client.Options{
		HostPort:  addr,
		Namespace: ns,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{},
		},
		Credentials: client.NewAPIKeyStaticCredentials(apiKey),
	})
	if err != nil {
		log.Fatalf("Failed to connect to Temporal Cloud: %v", err)
	}

	log.Println("✅ Client connected to Temporal Cloud")
	return c
}

func ctx() context.Context {
	return context.Background()
}

// ----------------- Basic Workflow Operations -----------------

func startWorkflow(c client.Client) {
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		log.Fatal("TEMPORAL_TASK_QUEUE is not set")
	}

	workflowID := fmt.Sprintf("%s%d", shared.WorkflowIDPrefix, time.Now().Unix())

	wo := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
		// Set a long execution timeout so workflow can wait for signal
		WorkflowExecutionTimeout: shared.DefaultWorkflowTimeout,
		WorkflowRunTimeout:       shared.DefaultWorkflowRunTimeout,
	}

	name := shared.DefaultWorkflowName

	log.Printf("🚀 Starting workflow: %s\n", workflowID)

	run, err := c.ExecuteWorkflow(ctx(), wo, shared.WorkflowName, name)
	if err != nil {
		log.Fatalf("Unable to start workflow: %v", err)
	}

	log.Printf("Workflow started. WorkflowID=%s RunID=%s\n", run.GetID(), run.GetRunID())
	log.Printf("✅ Workflow is running and waiting for signal...\n")
	log.Printf("\n👉 Next steps:\n")
	log.Printf("   1. Check status: go run client.go status %s\n", workflowID)
	log.Printf("   2. Send signal:  go run client.go signal %s NEW_NAME\n", workflowID)
	log.Printf("   3. Get result:   go run client.go get %s\n", workflowID)
}

func signalWorkflow(c client.Client) {
	// go run client.go signal <workflowID> <newName>
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run client.go signal <workflowID> <newName>")
	}

	workflowID := os.Args[2]
	newName := os.Args[3]

	err := c.SignalWorkflow(ctx(), workflowID, "", shared.SignalUpdateName, newName)
	if err != nil {
		log.Fatalf("Failed to send signal: %v", err)
	}

	log.Printf("📨 Sent Signal '%s' to workflow '%s' with value: %s\n",
		shared.SignalUpdateName, workflowID, newName)
	log.Printf("💡 Use 'go run client.go get %s' to get the workflow result\n", workflowID)
}

func startDelayWorkflow(c client.Client) {
	// go run client.go delay <minutes> [name]
	if len(os.Args) < 3 {
		log.Fatalf("Usage: go run client.go delay <minutes> [name]")
	}

	minutesStr := os.Args[2]
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil || minutes <= 0 {
		log.Fatalf("Invalid minutes: %s. Must be a positive integer", minutesStr)
	}

	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		log.Fatal("TEMPORAL_TASK_QUEUE is not set")
	}

	var name string
	if len(os.Args) >= 4 {
		name = os.Args[3]
	} else {
		name = shared.DefaultWorkflowName
	}

	workflowID := fmt.Sprintf("%s-delay-%d", shared.WorkflowIDPrefix, time.Now().Unix())

	wo := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
		// Set a long execution timeout so workflow can wait for delay and signal
		WorkflowExecutionTimeout: shared.DefaultWorkflowTimeout,
		WorkflowRunTimeout:       shared.DefaultWorkflowRunTimeout,
	}

	// DelayWorkflowInput describes the payload for DelayWorkflow
	type DelayWorkflowInput struct {
		DelayMinutes int
		Name         string
	}
	input := DelayWorkflowInput{
		DelayMinutes: minutes,
		Name:         name,
	}

	log.Printf("⏰ Starting delay workflow: %s\n", workflowID)
	log.Printf("   Will wait %d minute(s), then wait for signal...\n", minutes)

	run, err := c.ExecuteWorkflow(ctx(), wo, shared.DelayWorkflowName, input)
	if err != nil {
		log.Fatalf("Unable to start delay workflow: %v", err)
	}

	log.Printf("Workflow started. WorkflowID=%s RunID=%s\n", run.GetID(), run.GetRunID())
	log.Printf("✅ Workflow will wait %d minute(s), then wait for signal...\n", minutes)
	log.Printf("\n👉 Next steps:\n")
	log.Printf("   1. Wait %d minute(s) for the delay to complete\n", minutes)
	log.Printf("   2. Send signal:  go run client.go signal %s NEW_NAME\n", workflowID)
	log.Printf("   3. Get result:   go run client.go get %s\n", workflowID)
}

func getWorkflowResult(c client.Client) {
	// go run client.go get <workflowID>
	if len(os.Args) < 3 {
		log.Fatalf("Usage: go run client.go get <workflowID>")
	}

	workflowID := os.Args[2]

	log.Printf("⏳ Waiting for workflow '%s' to complete...\n", workflowID)

	run := c.GetWorkflow(ctx(), workflowID, "")
	var result string
	if err := run.Get(ctx(), &result); err != nil {
		log.Fatalf("Unable to get workflow result: %v", err)
	}
	log.Printf("🎉 Workflow completed. Result: %s\n", result)
}

func getWorkflowStatus(c client.Client) {
	// go run client.go status <workflowID>
	if len(os.Args) < 3 {
		log.Fatalf("Usage: go run client.go status <workflowID>")
	}

	workflowID := os.Args[2]

	desc, err := c.DescribeWorkflowExecution(ctx(), workflowID, "")
	if err != nil {
		log.Fatalf("Unable to get workflow status: %v", err)
	}

	status := desc.WorkflowExecutionInfo.Status.String()
	log.Printf("📊 Workflow Status: %s\n", status)
	log.Printf("   WorkflowID: %s\n", desc.WorkflowExecutionInfo.Execution.WorkflowId)
	log.Printf("   RunID: %s\n", desc.WorkflowExecutionInfo.Execution.RunId)

	if status == "RUNNING" {
		log.Printf("✅ Workflow is running and waiting for signal...\n")
		log.Printf("💡 Send signal: go run client.go signal %s NEW_NAME\n", workflowID)
	} else if status == "COMPLETED" {
		log.Printf("✅ Workflow has completed. Use 'get' command to see result.\n")
		log.Printf("💡 Get result: go run client.go get %s\n", workflowID)
	}
}

// ----------------- Schedule Command Entry Point -----------------

func handleScheduleCommand(c client.Client) {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: go run client.go schedule <create|list|pause|resume|delete|describe> [args...]")
	}

	subCommand := os.Args[2]
	scheduleClient := c.ScheduleClient()

	switch subCommand {
	case "create":
		createSchedule(scheduleClient)
	case "list":
		listSchedules(scheduleClient)
	case "pause":
		pauseSchedule(scheduleClient)
	case "resume":
		resumeSchedule(scheduleClient)
	case "delete":
		deleteSchedule(scheduleClient)
	case "describe":
		describeSchedule(scheduleClient)
	default:
		log.Fatalf("Unknown schedule command %q. Use: create | list | pause | resume | delete | describe", subCommand)
	}
}

// ----------------- Schedule: create / list / pause / resume / delete / describe -----------------

func createSchedule(sc client.ScheduleClient) {
	// go run client.go schedule create <scheduleID> <cron> [workflowID]
	if len(os.Args) < 5 {
		log.Fatalf("Usage: go run client.go schedule create <scheduleID> <cron> [workflowID]")
	}

	scheduleID := os.Args[3]
	cronExpr := os.Args[4]
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		log.Fatal("TEMPORAL_TASK_QUEUE is not set")
	}

	var workflowID string
	if len(os.Args) >= 6 {
		workflowID = os.Args[5]
	} else {
		workflowID = fmt.Sprintf("%s%d", shared.WorkflowIDPrefix, time.Now().Unix())
	}

	name := shared.DefaultWorkflowName

	// 1. Schedule trigger time rules (Cron)
	spec := client.ScheduleSpec{
		CronExpressions: []string{cronExpr},
	}

	// 2. Workflow action to execute
	// ScheduleWorkflow runs automatically and completes immediately (cron-style task)
	type ScheduleWorkflowInput struct {
		Name string
	}
	workflowInput := ScheduleWorkflowInput{
		Name: name,
	}

	action := &client.ScheduleWorkflowAction{
		ID: workflowID,
		// Run ScheduleWorkflow; it completes automatically without waiting for a signal
		Workflow:  shared.ScheduleWorkflowName,
		TaskQueue: taskQueue,
		Args:      []interface{}{workflowInput},
	}

	// 3. Complete ScheduleOptions
	opts := client.ScheduleOptions{
		ID:     scheduleID,
		Spec:   spec,
		Action: action,
		// Can add State / Policies / Memo / SearchAttributes if needed
	}

	handle, err := sc.Create(ctx(), opts)
	if err != nil {
		log.Fatalf("Failed to create schedule: %v\n", err)
	}

	log.Printf("✅ Schedule created: %s\n", scheduleID)
	log.Printf("   Cron: %s\n", cronExpr)
	log.Printf("   WorkflowID: %s\n", workflowID)
	log.Printf("💡 Use 'go run client.go schedule list' to see all schedules\n")
	_ = handle // Currently only creating, not using handle further
}

func listSchedules(sc client.ScheduleClient) {
	iter, err := sc.List(ctx(), client.ScheduleListOptions{})
	if err != nil {
		log.Fatalf("Failed to list schedules: %v", err)
	}

	log.Println("📋 Schedules:")
	count := 0
	for iter.HasNext() {
		entry, err := iter.Next()
		if err != nil {
			log.Fatalf("Failed to get schedule: %v", err)
		}
		count++
		// Only printing ID here, use Describe to see detailed status
		log.Printf("   %d. ID: %s\n", count, entry.ID)
	}

	if count == 0 {
		log.Println("   No schedules found")
	}
}

func pauseSchedule(sc client.ScheduleClient) {
	// go run client.go schedule pause <scheduleID>
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run client.go schedule pause <scheduleID>")
	}

	scheduleID := os.Args[3]
	handle := sc.GetHandle(ctx(), scheduleID)
	err := handle.Pause(ctx(), client.SchedulePauseOptions{})
	if err != nil {
		log.Fatalf("Failed to pause schedule: %v", err)
	}

	log.Printf("⏸️  Schedule paused: %s\n", scheduleID)
}

func resumeSchedule(sc client.ScheduleClient) {
	// go run client.go schedule resume <scheduleID>
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run client.go schedule resume <scheduleID>")
	}

	scheduleID := os.Args[3]
	handle := sc.GetHandle(ctx(), scheduleID)
	err := handle.Unpause(ctx(), client.ScheduleUnpauseOptions{})
	if err != nil {
		log.Fatalf("Failed to resume schedule: %v", err)
	}

	log.Printf("▶️  Schedule resumed: %s\n", scheduleID)
}

func deleteSchedule(sc client.ScheduleClient) {
	// go run client.go schedule delete <scheduleID>
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run client.go schedule delete <scheduleID>")
	}

	scheduleID := os.Args[3]
	handle := sc.GetHandle(ctx(), scheduleID)
	err := handle.Delete(ctx())
	if err != nil {
		log.Fatalf("Failed to delete schedule: %v", err)
	}

	log.Printf("🗑️  Schedule deleted: %s\n", scheduleID)
}

func describeSchedule(sc client.ScheduleClient) {
	// go run client.go schedule describe <scheduleID>
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run client.go schedule describe <scheduleID>")
	}

	scheduleID := os.Args[3]
	handle := sc.GetHandle(ctx(), scheduleID)
	desc, err := handle.Describe(ctx())
	if err != nil {
		log.Fatalf("Failed to describe schedule: %v", err)
	}

	log.Printf("📊 Schedule: %s\n", scheduleID)

	state := "Active"
	if desc.Schedule.State.Paused {
		state = "Paused"
	}
	log.Printf("   State: %s\n", state)

	if len(desc.Schedule.Spec.CronExpressions) > 0 {
		log.Printf("   Cron: %s\n", desc.Schedule.Spec.CronExpressions[0])
	}

	if wfAction, ok := desc.Schedule.Action.(*client.ScheduleWorkflowAction); ok {
		log.Printf("   WorkflowID: %s\n", wfAction.ID)
		log.Printf("   TaskQueue: %s\n", wfAction.TaskQueue)
	}
}
