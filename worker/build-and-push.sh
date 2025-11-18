#!/usr/bin/env bash
set -e

# Script to build Docker image and push to AWS ECR
# Usage: ./build-and-push.sh [image-tag]
# Example: ./build-and-push.sh v1.0.0

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables from parent directory if .env exists
if [ -f "../.env" ]; then
  echo "Loading .env from ../.env"
  export $(grep -v '^#' "../.env" | xargs)
elif [ -f ".env" ]; then
  echo "Loading .env from .env"
  export $(grep -v '^#' ".env" | xargs)
fi

# Get image tag from argument or use default
IMAGE_TAG="${1:-latest}"
PROJECT_NAME="${PROJECT_NAME:-temporal-demo}"
AWS_REGION="${AWS_REGION:-us-east-1}"

# ECR repository name (should match what's in Terraform)
ECR_REPO_NAME="${PROJECT_NAME}-worker"

# Use terraform profile if available, otherwise use default
AWS_PROFILE="${AWS_PROFILE:-terraform}"

# Get AWS account ID
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --profile "${AWS_PROFILE}" --query Account --output text)
if [ -z "$AWS_ACCOUNT_ID" ]; then
  echo "❌ Failed to get AWS account ID. Make sure AWS CLI is configured."
  echo "   Try: aws sso login --profile terraform"
  exit 1
fi

# ECR repository URL
ECR_REPO_URL="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPO_NAME}"

echo "📦 Building Docker image..."
echo "   Repository: ${ECR_REPO_URL}"
echo "   Tag: ${IMAGE_TAG}"
echo "   Platform: linux/amd64 (for ECS Fargate compatibility)"
echo ""

# Build the Docker image for linux/amd64 platform (required for ECS Fargate)
docker build --platform=linux/amd64 -t "${ECR_REPO_NAME}:${IMAGE_TAG}" .

# Also tag with full ECR URL
docker tag "${ECR_REPO_NAME}:${IMAGE_TAG}" "${ECR_REPO_URL}:${IMAGE_TAG}"

echo ""
echo "🔐 Logging in to ECR..."
aws ecr get-login-password --profile "${AWS_PROFILE}" --region "${AWS_REGION}" | docker login --username AWS --password-stdin "${ECR_REPO_URL}"

echo ""
echo "📤 Pushing image to ECR..."
docker push "${ECR_REPO_URL}:${IMAGE_TAG}"

echo ""
echo "✅ Successfully pushed ${ECR_REPO_URL}:${IMAGE_TAG}"
echo ""
echo "💡 Next steps:"
echo "   1. Update Terraform variable worker_image_tag to '${IMAGE_TAG}' (or use latest)"
echo "   2. Run: cd ../terraform && ./run-terraform.sh apply"
echo "   3. Or update ECS service to use the new image tag"

