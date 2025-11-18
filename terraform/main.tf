# main.tf

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# AWS provider configuration
provider "aws" {
  region = var.aws_region
}

########################
# Network (VPC / Subnet)
########################

# VPC for hosting all network resources
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project_name}-vpc"
  }
}

# Public subnet where ECS Fargate tasks will run
# Must assign public IP so workers can reach Temporal Cloud
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = "${var.aws_region}a"

  tags = {
    Name = "${var.project_name}-public-subnet"
  }
}

# Internet Gateway to allow outbound internet access
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-igw"
  }
}

# Route table that sends all outbound traffic (0.0.0.0/0)
# through the Internet Gateway
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "${var.project_name}-public-rt"
  }
}

# Associate the public route table with the public subnet
resource "aws_route_table_association" "public_assoc" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

#################
# Security Group
#################

# Security group for the Temporal Worker running on ECS
# - Outbound: allow all traffic (to reach Temporal Cloud API)
# - Inbound: restricted to traffic from within the VPC
resource "aws_security_group" "worker" {
  name        = "${var.project_name}-worker-sg"
  description = "Security group for Temporal worker"
  vpc_id      = aws_vpc.main.id

  # Allow all outbound traffic (needed to access Temporal Cloud over HTTPS)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow inbound traffic from within the VPC CIDR (for internal communication / debugging)
  ingress {
    description = "Allow all inbound from VPC"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  tags = {
    Name = "${var.project_name}-worker-sg"
  }
}

###############
# ECR Repo
###############

# ECR repository to store the Temporal Worker Docker image
resource "aws_ecr_repository" "worker" {
  name                 = "${var.project_name}-worker"
  image_tag_mutability = "MUTABLE"
  force_delete         = true  # Allow deletion even if repository contains images

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "${var.project_name}-worker-ecr"
  }
}

###########################
# CloudWatch Log Group
###########################

# CloudWatch log group for worker logs (workflow/activities logs, errors, etc.)
resource "aws_cloudwatch_log_group" "worker" {
  name              = "/ecs/${var.project_name}-worker"
  retention_in_days = 7
}

#######################
# ECS Cluster (Fargate)
#######################

# ECS Cluster for running the Temporal Worker service on Fargate
resource "aws_ecs_cluster" "worker" {
  name = "${var.project_name}-worker-cluster"
}

#######################
# IAM Roles for ECS
#######################

# Trust policy: ECS tasks can assume this role
data "aws_iam_policy_document" "ecs_task_execution_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# Execution role: ECS tasks use this role to pull images from ECR and write logs
resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "${var.project_name}-ecs-execution-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_execution_assume_role.json
}

# Attach AWS-managed policy for ECS task execution
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Task role: assumed by the worker container itself
# Use this for granting access to AWS services like S3/SSM/DynamoDB later
resource "aws_iam_role" "ecs_task_role" {
  name               = "${var.project_name}-ecs-task-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_execution_assume_role.json
}

###########################
# ECS Task Definition
###########################

# Local value for container name, reused in the task definition
locals {
  worker_container_name = "${var.project_name}-worker"
}

# ECS Task Definition for the Temporal Worker container
# Includes:
# - Image from ECR
# - CPU/Memory settings
# - Environment variables for Temporal Cloud connection
# - CloudWatch logging configuration
resource "aws_ecs_task_definition" "worker" {
  family                   = "${var.project_name}-worker-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name      = local.worker_container_name
      image     = "${aws_ecr_repository.worker.repository_url}:${var.worker_image_tag}"
      essential = true

      # Environment variables used by the worker to connect to Temporal Cloud
      environment = [
        {
          name  = "TEMPORAL_ADDRESS"
          value = var.temporal_address
        },
        {
          name  = "TEMPORAL_NAMESPACE"
          value = var.temporal_namespace
        },
        {
          name  = "TEMPORAL_API_KEY"
          value = var.temporal_api_key
        },
        {
          name  = "TEMPORAL_TASK_QUEUE"
          value = "${var.project_name}-queue"
        }
      ]

      # Send container logs to CloudWatch Logs
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.worker.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }
    }
  ])
}

###########################
# ECS Service (Worker)
###########################

# ECS Service ensures the worker always keeps running
# - Uses Fargate launch type
# - Runs inside the public subnet with a public IP
# - Uses the worker security group for outbound access to Temporal Cloud
resource "aws_ecs_service" "worker" {
  name            = "${var.project_name}-worker-svc"
  cluster         = aws_ecs_cluster.worker.id
  task_definition = aws_ecs_task_definition.worker.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = [aws_subnet.public.id]
    security_groups  = [aws_security_group.worker.id]
    assign_public_ip = true
  }

  lifecycle {
    # Ignore manual changes to desired_count (useful if you scale temporarily)
    ignore_changes = [desired_count]
  }

  depends_on = [
    aws_internet_gateway.igw
  ]
}



