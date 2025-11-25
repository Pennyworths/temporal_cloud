# Temporal Cloud Demo Project

A complete example project demonstrating how to use Temporal Cloud, showcasing core features including Workflow, Activity, Signal, and Schedule.

---

## 📚 Table of Contents

- [What is Temporal Cloud?](#what-is-temporal-cloud)
- [Core Concepts](#core-concepts)
  - [Workflow](#workflow)
  - [Activity](#activity)
  - [Signal](#signal)
  - [Schedule](#schedule)
- [Quick Start](#quick-start)
- [Deployment Steps](#deployment-steps)
- [Usage Guide](#usage-guide)
- [Complete Examples](#complete-examples)

---

## What is Temporal Cloud?

**Temporal Cloud** is a managed service provided by Temporal Technologies for building reliable and scalable distributed applications.

### Core Value

- ✅ **Reliability**: Ensures workflows execute reliably even during system failures
- ✅ **Observability**: Complete execution history and state tracking
- ✅ **Scalability**: Automatically handles high concurrency and load
- ✅ **Persistence**: Workflow state is persisted, unaffected by service restarts

### Typical Use Cases

- Order processing workflows
- Data synchronization and backup
- Scheduled task execution
- Long-running business processes
- Tasks requiring retry and fault tolerance

---

## Core Concepts

### Workflow

**Workflow** is the orchestration layer for business logic, defining the execution order and flow of tasks.

#### Characteristics:
- 🔄 **Deterministic**: Same inputs always produce the same results
- 💾 **Persistent**: Execution state is saved in Temporal
- 🔁 **Recoverable**: Can resume from the last state even if the worker crashes
- ⏸️ **Pausable**: Can wait for external events (such as Signals)

#### Example:
```go
func HelloWorkflow(ctx workflow.Context, name string) (string, error) {
    // Wait for signal
    signalChan := workflow.GetSignalChannel(ctx, "update-name")
    var newName string
    signalChan.Receive(ctx, &newName)
    
    // Execute logic
    result := fmt.Sprintf("Hello, %s!", newName)
    return result, nil
}
```

### Activity

**Activity** is a function that performs actual work, such as calling external APIs, operating databases, etc.

#### Characteristics:
- 🌐 **Non-deterministic**: Can call external services, read files, etc.
- ⏱️ **Timeout Control**: Can set timeout duration
- 🔄 **Auto-retry**: Automatically retries on failure
- 🔌 **Cancellable**: Can be cancelled

#### Example:
```go
func SayHello(ctx context.Context, name string) (string, error) {
    // Perform actual work (call API, operate database, etc.)
    return "Hello, " + name + " from Temporal Worker on AWS!", nil
}
```

### Signal

**Signal** is a mechanism for sending asynchronous messages from external sources to running Workflows.

#### How Workflow, Activity, and Signal Work Together

- **Workflow orchestrates the entire business process.** It determines when to run Activities and when to pause for external input.
- **Activities perform the actual work.** A workflow calls an activity whenever it needs to execute nondeterministic operations (API calls, DB writes, etc.).
- **Signals feed the workflow with live updates.** While a workflow is waiting on `workflow.GetSignalChannel`, external systems can use the client CLI (or API) to deliver new data that unblocks the workflow and lets it proceed.

In short: **Workflow = conductor**, **Activity = musicians**, **Signal = request from the audience**. The workflow coordinates, activities do the heavy lifting, and signals allow outside systems to nudge the workflow mid-flight.

#### Use Cases:
- 📨 **External Trigger**: Allow Workflow to respond to external events
- 🔄 **State Update**: Update Workflow execution parameters
- ⏸️ **Resume Execution**: Allow waiting Workflows to continue execution

#### Workflow:
```
Workflow starts
    ↓
Workflow task is placed on the Task Queue (worker polls and picks it up)
    ↓
Wait for Signal (blocked)
    ↓
External sends Signal
    ↓
Workflow receives Signal
    ↓
Workflow invokes Activities (e.g., SayHello) to perform real work
    ↓
Continue execution and complete
```

#### Workflow Execution Modes

1. **Auto-complete Workflow (no signal required)**
```
Workflow starts (autoStart = true, e.g., ScheduleWorkflow)
    ↓
Workflow invokes Activities directly
    ↓
Workflow completes immediately without waiting for a Signal
```

2. **Signal-driven Workflow**
```
Workflow starts (autoStart = false, e.g., HelloWorkflow via CLI)
    ↓
Workflow pauses and waits for Signal
    ↓
External system sends Signal (client CLI, API, etc.)
    ↓
Workflow receives Signal → invokes Activities → completes
```

Under the hood the workflow code in `worker/workflow.go` is:

```go
signalChan := workflow.GetSignalChannel(ctx, shared.SignalUpdateName)
signalChan.Receive(ctx, &newName)
result := fmt.Sprintf("Hello, %s from Temporal Worker on AWS!", newName)
return result, nil
```

And the CLI sends the signal with:

```go
err := c.SignalWorkflow(ctx(), workflowID, "", shared.SignalUpdateName, newName)
```

### Task Queue

**Task Queue** is the bridge between the client (that schedules work) and the worker (that executes it). In this project:

- The CLI reads `TEMPORAL_TASK_QUEUE` from `.env` and sets it in `client.StartWorkflowOptions` or `ScheduleWorkflowAction`. Every workflow run is therefore enqueued on the same Task Queue.
- `worker/main.go` creates a worker with the identical Task Queue name:
  ```go
  w := worker.New(c, taskQueue, worker.Options{})
  ```
  That worker polls the queue, picks up workflow or activity tasks, and executes the registered handlers.

Temporal ensures horizontal scalability: you can run multiple workers pointing to the same Task Queue, and tasks will be load-balanced across them automatically.

#### Example:
```bash
# Start workflow
go run client.go start
# Output: WorkflowID=hello-workflow-123

# Send signal
go run client.go signal hello-workflow-123 Alice

# Fetch result
go run client.go get hello-workflow-123
# Output:
# ⏳ Waiting for workflow 'hello-workflow-123' to complete...
# 🎉 Workflow completed. Result: Hello, Alice from Temporal Worker on AWS!
```

### Schedule

**Schedule** is a mechanism for automatically executing Workflows according to a time schedule, similar to Cron jobs.

#### Characteristics:
- ⏰ **Scheduled Trigger**: Automatically triggers based on Cron expressions
- 💾 **Persistent**: Saved in Temporal, unaffected by service restarts
- 🎛️ **Manageable**: Can pause, resume, and delete
- 📊 **Observable**: Complete execution history

#### Workflow:
```
Create Schedule (Cron: "0 * * * *")
    ↓
Automatically triggers every hour
    ↓
Automatically executes Workflow
    ↓
Automatically completes (no manual operation needed)
```

Behind the scenes, the CLI builds a `ScheduleWorkflowAction` and Temporal Cloud executes the registered workflow:

```go
action := &client.ScheduleWorkflowAction{
    ID:        workflowID,
    Workflow:  shared.ScheduleWorkflowName,
    TaskQueue: taskQueue,
    Args:      []interface{}{ScheduleWorkflowInput{Name: name}},
}
```

And the workflow itself resides in `worker/workflow.go`:

```go
func ScheduleWorkflow(ctx workflow.Context, input ScheduleWorkflowInput) (string, error) {
    name := input.Name
    if name == "" {
        name = shared.DefaultWorkflowName
    }
    result := fmt.Sprintf("Hello, %s from Temporal Worker on AWS!", name)
    return result, nil
}
```

#### Example:
```bash
# Create a schedule that runs every hour
go run client.go schedule create hourly-task "0 * * * *"

# Schedule will automatically:
# - Trigger every hour
# - Execute Workflow
# - Complete automatically
```

---

## Quick Start

### Prerequisites

1. **AWS CLI configured**
   ```bash
   aws configure
   # Or set environment variables
   export AWS_ACCESS_KEY_ID=your-key
   export AWS_SECRET_ACCESS_KEY=your-secret
   export AWS_DEFAULT_REGION=us-east-1
   ```

2. **Docker installed and running**

3. **Go 1.23+ installed**

4. **Terraform 1.5+ installed**

5. **Temporal Cloud account**
   - Create a Namespace on [Temporal Cloud](https://cloud.temporal.io)
   - Get API Key and Endpoint

### How the code connects to Temporal Cloud

Both the CLI and the worker read the connection settings from `.env` and build identical Temporal clients:

```go
clientOptions := client.Options{
    HostPort:  os.Getenv("TEMPORAL_ADDRESS"),
    Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
    ConnectionOptions: client.ConnectionOptions{
        TLS: &tls.Config{},
    },
    Credentials: client.NewAPIKeyStaticCredentials(os.Getenv("TEMPORAL_API_KEY")),
}

c, err := client.Dial(clientOptions)
```

Once the client is ready, the worker registers workflows/activities on the shared Task Queue:

```go
w := worker.New(c, taskQueue, worker.Options{})
w.RegisterWorkflow(HelloWorkflow)
w.RegisterWorkflow(ScheduleWorkflow)
w.RegisterWorkflow(DelayWorkflow)
w.RegisterActivity(SayHello)
w.Run(worker.InterruptCh())
```

The CLI uses the same client handle when it executes `ExecuteWorkflow`, `SignalWorkflow`, `GetWorkflow`, or any of the schedule operations, so every command is routed through Temporal Cloud using the credentials above.

---

## Deployment Steps

### Step 1: Initialize Project

```bash
# In project root directory
cd your_workspace/temporal_cloud
go mod tidy
```

### Step 2: Configure Environment Variables

Create a Namespace on Temporal Cloud, get the API Key and Endpoint, then create a `.env` file:

```bash
# Create .env file
PROJECT_NAME=temporal-demo
AWS_REGION=us-east-1
TEMPORAL_ADDRESS=your-namespace.xyz.temporal.io:7233
TEMPORAL_NAMESPACE=your-namespace
TEMPORAL_API_KEY=your-api-key
TEMPORAL_TASK_QUEUE=temporal-demo-queue
```

### Step 3: Create AWS Infrastructure

```bash
cd terraform

# Initialize Terraform (first time only)
terraform init

# Create all AWS resources
./run-terraform.sh apply

# Confirm creation (type yes)
```

**This will create:**
- VPC, subnets, Internet Gateway
- ECR repository
- ECS cluster and service
- IAM roles
- CloudWatch log group
- Security groups, etc.

### Step 4: Build and Push Docker Image

```bash
cd ../worker

# Build and push image (using linux/amd64 platform)
./build-and-push.sh latest
```

**This will:**
- Build Docker image (Go 1.23, linux/amd64)
- Login to ECR
- Push image to ECR repository

### Step 5: Wait for ECS Service to Start

```bash
# Check ECS task status (optional)
aws ecs list-tasks \
  --cluster temporal-demo-worker-cluster \
  --service-name temporal-demo-worker-svc \
  --region us-east-1

# View CloudWatch logs (optional)
aws logs tail /ecs/temporal-demo-worker --follow --region us-east-1
```

**Wait for task status to become "Running"**

### Step 6: Start Worker (Local Testing)

Before running the client, you need to start the worker to process workflow tasks.

**Option 1: Run Worker Locally (Recommended for Testing)**

```bash
cd worker
go run .
```

The worker will run continuously, listening to the task queue. Keep this terminal window open.

**Option 2: Use Worker on ECS**

If already deployed to ECS, ensure the ECS service is running.

---

## Usage Guide

### Basic Workflow Operations

Every CLI command maps directly to a Temporal SDK call. For example:

- `go run client.go start` → `c.ExecuteWorkflow(ctx(), startOptions, shared.WorkflowName, shared.DefaultWorkflowName)`
- `go run client.go signal <WorkflowID> <NewName>` → `c.SignalWorkflow(ctx(), workflowID, "", shared.SignalUpdateName, newName)`
- `go run client.go get <WorkflowID>` → `c.GetWorkflow(ctx(), workflowID, "").Get(ctx(), &result)`

#### 1. Start Workflow

Start a workflow. It will wait for a signal before continuing execution.

```bash
cd client
go run client.go start
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
🚀 Starting workflow: hello-workflow-1763499702
Workflow started. WorkflowID=hello-workflow-1763499702 RunID=019a98c5-fa1e-7920-9b0b-8a254e4a238b
✅ Workflow is running and waiting for signal...

👉 Next steps:
   1. Check status: go run client.go status hello-workflow-1763499702
   2. Send signal:  go run client.go signal hello-workflow-1763499702 NEW_NAME
   3. Get result:   go run client.go get hello-workflow-1763499702
```

**Important:** Remember the `WorkflowID` from the output, as you'll need it for subsequent operations.

#### 2. Check Workflow Status

```bash
go run client.go status <WorkflowID>
```

**Example:**
```bash
go run client.go status hello-workflow-1763499702
```

**Output Example:**
```
📊 Workflow Status: RUNNING
   WorkflowID: hello-workflow-1763499702
   RunID: 019a98c5-fa1e-7920-9b0b-8a254e4a238b
✅ Workflow is running and waiting for signal...
💡 Send signal: go run client.go signal hello-workflow-1763499702 NEW_NAME
```

**Status Description:**
- `RUNNING`: Workflow is running and waiting for a signal
- `COMPLETED`: Workflow has completed

#### 3. Send Signal

Send a signal to a running workflow to trigger it to continue execution.

```bash
go run client.go signal <WorkflowID> <NewName>
```

**Parameters:**
- `<WorkflowID>`: Workflow ID (obtained from the `start` command)
- `<NewName>`: New name (can be any string)

**Example:**
```bash
go run client.go signal hello-workflow-1763499702 Alice
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
📨 Sent Signal 'update-name' to workflow 'hello-workflow-1763499702' with value: Alice
💡 Use 'go run client.go get hello-workflow-1763499702' to get the workflow result
```

#### 4. Get Workflow Result

Wait for the workflow to complete and get the final result.

```bash
go run client.go get <WorkflowID>
```

**Example:**
```bash
go run client.go get hello-workflow-1763499702
```

**Output Example:**
```
⏳ Waiting for workflow 'hello-workflow-1763499702' to complete...
🎉 Workflow completed. Result: Hello, Alice from Temporal Worker on AWS!
```

---

### Schedule Feature

Schedule allows you to create scheduled tasks that automatically execute workflows based on Cron expressions.

#### 1. Create Schedule

Create a scheduled task that automatically executes a workflow based on a specified Cron expression.

```bash
go run client.go schedule create <ScheduleID> <CronExpression> [WorkflowID]
```

**Parameters:**
- `<ScheduleID>`: Unique identifier for the scheduled task
- `<CronExpression>`: Cron expression (e.g., `"0 * * * *"` means execute every hour)
- `[WorkflowID]`: Optional, workflow ID (if not provided, will be auto-generated)

**Cron Expression Examples:**
- `"0 * * * *"` - Execute every hour
- `"0 0 * * *"` - Execute daily at midnight
- `"*/5 * * * *"` - Execute every 5 minutes
- `"0 9 * * 1"` - Execute every Monday at 9 AM

**Examples:**
```bash
# Create a scheduled task that runs every hour
go run client.go schedule create hourly-task "0 * * * *"

# Create a daily scheduled task
go run client.go schedule create daily-task "0 0 * * *"
```

**Important:** Cron expressions must be wrapped in quotes to prevent shell from parsing wildcards.

**Output Example:**
```
✅ Client connected to Temporal Cloud
✅ Schedule created: hourly-task
   Cron: 0 * * * *
   WorkflowID: hello-workflow-1763499702
💡 Use 'go run client.go schedule list' to see all schedules
```

#### 2. List Schedules

```bash
go run client.go schedule list
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
📋 Schedules:
   1. ID: hourly-task
   2. ID: daily-task
```

#### 3. Describe Schedule

```bash
go run client.go schedule describe <ScheduleID>
```

**Example:**
```bash
go run client.go schedule describe hourly-task
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
📊 Schedule: hourly-task
   State: Active
   Cron: 0 * * * *
   WorkflowID: hello-workflow-1763499702
   TaskQueue: temporal-demo-queue
```

#### 4. Pause Schedule

Pause a scheduled task to stop automatic execution.

```bash
go run client.go schedule pause <ScheduleID>
```

**Example:**
```bash
go run client.go schedule pause hourly-task
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
⏸️  Schedule paused: hourly-task
```

#### 5. Resume Schedule

Resume a paused scheduled task.

```bash
go run client.go schedule resume <ScheduleID>
```

**Example:**
```bash
go run client.go schedule resume hourly-task
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
▶️  Schedule resumed: hourly-task
```

#### 6. Delete Schedule

Delete a scheduled task.

```bash
go run client.go schedule delete <ScheduleID>
```

**Example:**
```bash
go run client.go schedule delete hourly-task
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
🗑️  Schedule deleted: hourly-task
```

---

### Delay Feature

The delay feature allows you to start a workflow after a specified time and wait for a signal.

It sends a structured payload into `DelayWorkflow`:

```go
type DelayWorkflowInput struct {
    DelayMinutes int
    Name         string
}
```

Workflow logic:

```go
workflow.Sleep(ctx, time.Duration(delayMinutes)*time.Minute)
signalChan := workflow.GetSignalChannel(ctx, shared.SignalUpdateName)
signalChan.Receive(ctx, &newName)
```

```bash
go run client.go delay <minutes> [name]
```

**Examples:**
```bash
# Start waiting for signal after 5 minutes
go run client.go delay 5

# Start waiting for signal after 10 minutes, using name Alice
go run client.go delay 10 Alice
```

**Workflow:**
1. Workflow starts immediately
2. Wait for the specified number of minutes (e.g., 5 minutes)
3. After delay ends, start waiting for signal
4. Send signal: `go run client.go signal <workflowID> NEW_NAME`
5. Workflow completes after receiving signal

---

## Complete Examples

### Example 1: Basic Workflow + Signal Flow

```bash
# 1. Start Worker (in one terminal window)
cd worker
go run .

# 2. Start workflow (in another terminal window)
cd client
go run client.go start
# Output: WorkflowID=hello-workflow-1234567890

# 3. Check status
go run client.go status hello-workflow-1234567890
# Output: Status: RUNNING

# 4. Send signal
go run client.go signal hello-workflow-1234567890 Bob

# 5. Get result
go run client.go get hello-workflow-1234567890
# Output: Hello, Bob from Temporal Worker on AWS!
```

### Example 2: Using Schedule

```bash
# 1. Create a scheduled task (runs every hour)
go run client.go schedule create my-hourly-task "0 * * * *"

# 2. List all scheduled tasks
go run client.go schedule list

# 3. Describe a scheduled task
go run client.go schedule describe my-hourly-task

# 4. Pause a scheduled task
go run client.go schedule pause my-hourly-task

# 5. Resume a scheduled task
go run client.go schedule resume my-hourly-task

# 6. Delete a scheduled task
go run client.go schedule delete my-hourly-task
```

### Example 3: Delay Execution

```bash
# 1. Start delay workflow (wait for signal after 5 minutes)
go run client.go delay 5
# Output: WorkflowID=hello-workflow-delay-1234567890

# 2. Wait 5 minutes...

# 3. Send signal
go run client.go signal hello-workflow-delay-1234567890 Alice

# 4. Get result
go run client.go get hello-workflow-delay-1234567890
```

---

## Command Reference

### Workflow Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `start` | `go run client.go start` | Start a workflow, returns immediately |
| `status` | `go run client.go status <WorkflowID>` | Check workflow status |
| `signal` | `go run client.go signal <WorkflowID> <NewName>` | Send signal to workflow |
| `get` | `go run client.go get <WorkflowID>` | Get workflow result (waits for completion) |

### Schedule Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `create` | `go run client.go schedule create <ScheduleID> <Cron> [WorkflowID]` | Create a scheduled task |
| `list` | `go run client.go schedule list` | List all scheduled tasks |
| `describe` | `go run client.go schedule describe <ScheduleID>` | Describe a scheduled task |
| `pause` | `go run client.go schedule pause <ScheduleID>` | Pause a scheduled task |
| `resume` | `go run client.go schedule resume <ScheduleID>` | Resume a scheduled task |
| `delete` | `go run client.go schedule delete <ScheduleID>` | Delete a scheduled task |

### Delay Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `delay` | `go run client.go delay <minutes> [name]` | Delay workflow execution |

---

## Verification and Monitoring (Optional)

- **Temporal Web UI**: Log in to Temporal Cloud and check workflow execution history
- **CloudWatch Logs**: Confirm worker is running properly
  ```bash
  aws logs tail /ecs/temporal-demo-worker --follow --region us-east-1
  ```
- **ECS Task Status**: Check if worker tasks are running
  ```bash
  aws ecs list-tasks \
    --cluster temporal-demo-worker-cluster \
    --service-name temporal-demo-worker-svc \
    --region us-east-1
  ```

---

## Important Notes

1. **`.env` file**: Must contain all required configurations
2. **Task Queue name**: Client and Worker must use the same name
3. **Worker must be running**: Both workflows and scheduled tasks require a worker to process
4. **WorkflowID**: Remember each workflow's ID for subsequent operations
5. **Cron expressions**: Use standard Cron format, pay attention to timezone settings
6. **Image platform**: Must build `linux/amd64` platform (required by ECS Fargate)
7. **AWS credentials**: Ensure AWS CLI is properly configured
8. **ECR images**: First deployment requires creating ECR repository first, then pushing image

---

## Project Structure

```
temporal_cloud/
├── client/              # Client code (start workflow, send signal, etc.)
│   └── client.go
├── worker/              # Worker code (execute workflow and activity)
│   ├── main.go         # Worker main program
│   ├── workflow.go     # Workflow definitions
│   ├── activity.go      # Activity definitions
│   └── Dockerfile       # Docker image definition
├── shared/             # Shared code
│   └── constants.go     # Constant definitions
├── terraform/          # Terraform configuration
│   ├── main.tf         # Main resource definitions
│   ├── variables.tf    # Variable definitions
│   └── output.tf       # Output definitions
└── README.md           # This document
```

---

## Summary

This project demonstrates the core features of Temporal Cloud:

- ✅ **Workflow**: Define business logic flow
- ✅ **Activity**: Execute specific work
- ✅ **Signal**: External trigger to continue workflow execution
- ✅ **Schedule**: Automatically execute workflows on a schedule

With these features, you can build reliable and scalable distributed applications!
