#!/usr/bin/env bash
set -e


SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"


# Try to load .env from current directory or parent directory
if [ -f "${SCRIPT_DIR}/.env" ]; then
  echo "Loading .env from ${SCRIPT_DIR}/.env"
  export $(grep -v '^#' "${SCRIPT_DIR}/.env" | xargs)
elif [ -f "${SCRIPT_DIR}/../.env" ]; then
  echo "Loading .env from ${SCRIPT_DIR}/../.env"
  export $(grep -v '^#' "${SCRIPT_DIR}/../.env" | xargs)
else
  echo "No .env file found. Looking in ${SCRIPT_DIR}/.env or ${SCRIPT_DIR}/../.env"
fi

export TF_VAR_project_name="$PROJECT_NAME"
export TF_VAR_aws_region="$AWS_REGION"
export TF_VAR_temporal_address="$TEMPORAL_ADDRESS"
export TF_VAR_temporal_namespace="$TEMPORAL_NAMESPACE"
export TF_VAR_temporal_api_key="$TEMPORAL_API_KEY"
export TF_VAR_temporal_task_queue="$TEMPORAL_TASK_QUEUE"
export TF_VAR_worker_image_tag="latest"   


# Script is already in the terraform directory
TF_DIR="${SCRIPT_DIR}"
cd "$TF_DIR"

echo "Running terraform in $TF_DIR"
terraform "$@"
