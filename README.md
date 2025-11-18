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
cd /Users/yangtianfan/Desktop/temporal_cloud
go mod tidy
```

### Step 2: create a NameSpace on temporal cloud , get API KEY and region endpoint 
on temporal cloud.  Store environment Variable in .env
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

### Step 6: Test the client

```bash
cd ../clent

# Run client to trigger workflow
go run clent.go
```

**Expected output:**
```
✅ Client connected to Temporal Cloud
🚀 Started workflow. ID=hello-workflow-1234567890 RunID=...
🎉 Workflow completed. Result: Hello, Tianfan from Temporal Worker on AWS!
```

### Step 6: Verification (optional)

- **View in Temporal Web UI**: Log in to Temporal Cloud and check workflow execution history
- **Check CloudWatch logs**: Confirm worker is running properly
---

```

## Important Notes

1. **`.env` file**: Must contain all required configurations
2. **Task Queue name**: Client and Worker must use the same name
3. **Image platform**: Must build `linux/amd64` platform (required by ECS Fargate)
4. **AWS credentials**: Ensure AWS CLI is properly configured
5. **ECR images**: First deployment requires creating ECR repository first, then pushing image
