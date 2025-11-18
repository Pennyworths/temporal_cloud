variable "aws_region" {
  description = "AWS region to deploy resources into"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name prefix for all resources"
  type        = string
  default     = "temporal-demo"
}

variable "temporal_namespace" {
  description = "Temporal Cloud namespace"
  type        = string
}

variable "temporal_address" {
  description = "Temporal Cloud address (host:port)"
  type        = string
}

variable "worker_image_tag" {
  description = "Docker image tag for the Temporal worker"
  type        = string
  default     = "latest"
}

variable "temporal_api_key" {
  description = "Temporal Cloud API key"
  type        = string
  sensitive   = true
}