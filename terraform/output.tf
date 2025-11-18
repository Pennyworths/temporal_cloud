output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnet_id" {
  description = "Public subnet ID"
  value       = aws_subnet.public.id
}

output "worker_ecr_repository_url" {
  description = "ECR repository URL for Temporal worker"
  value       = aws_ecr_repository.worker.repository_url
}

output "ecs_cluster_name" {
  description = "ECS cluster name for Temporal worker"
  value       = aws_ecs_cluster.worker.name
}

output "ecs_service_name" {
  description = "ECS service name running the worker"
  value       = aws_ecs_service.worker.name
}

output "cloudwatch_log_group" {
  description = "CloudWatch log group for worker logs"
  value       = aws_cloudwatch_log_group.worker.name
}

