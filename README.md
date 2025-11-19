## Prerequisites
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

---

## Deployment Steps

### Step 1: Initialize Go modules (if needed)

```bash
# In project root directory
cd your_workspace/temporal_cloud
go mod tidy
```

### Step 2: create a NameSpace on temporal cloud , get API KEY and region endpoint 
###on temporal cloud.  Store environment Variable in .env
   ```bash
   # Create .env file
   PROJECT_NAME=temporal-demo
   AWS_REGION=us-east-1
   TEMPORAL_ADDRESS=your-namespace.xyz.temporal.io:7233
   TEMPORAL_NAMESPACE=your-namespace
   TEMPORAL_API_KEY=your-api-key
   TEMPORAL_TASK_QUEUE=temporal-demo-queue
   ```

### Step 3: Create AWS infrastructure with Terraform

```bash
cd terraform

# Initialize Terraform (first time only)
terraform init

# Apply configuration to create all AWS resources
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

### Step 4: Build and push Docker image to ECR

```bash
cd ../worker

# Build and push image (using linux/amd64 platform)
./build-and-push.sh latest
```

**This will:**
- Build Docker image (Go 1.23, linux/amd64)
- Login to ECR
- Push image to ECR repository

### Step 5: Wait for ECS service to start

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

### Step 6: Run Worker (Local Testing)

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

#### 1. Start Workflow

Start a workflow. The workflow will wait for a signal before continuing execution.

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

Check the current status of a workflow.

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

# Create a daily scheduled task with a specific WorkflowID
go run client.go schedule create daily-task "0 0 * * *" my-workflow-123
```

**Output Example:**
```
✅ Client connected to Temporal Cloud
✅ Schedule created: hourly-task
   Cron: 0 * * * *
   WorkflowID: hello-workflow-1763499702
💡 Use 'go run client.go schedule list' to see all schedules
```

#### 2. List Schedules

View all created scheduled tasks.

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

View detailed information about a specific scheduled task.

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

### Complete Usage Examples

#### Example 1: Basic Workflow + Signal Flow

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

#### Example 2: Using Schedule

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

### Step 7: Verification and Monitoring (Optional)

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

