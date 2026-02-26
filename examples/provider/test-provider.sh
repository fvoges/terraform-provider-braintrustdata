#!/bin/bash
set -euo pipefail

echo "=========================================="
echo "Testing Braintrust Terraform Provider"
echo "=========================================="
echo ""

if [ -z "${BRAINTRUST_API_KEY:-}" ]; then
  echo "BRAINTRUST_API_KEY environment variable not set"
  echo "Set it with: export BRAINTRUST_API_KEY=\"sk-your-key\""
  exit 1
fi

echo "BRAINTRUST_API_KEY is set"

if [ -z "${BRAINTRUST_ORG_ID:-}" ]; then
  echo "BRAINTRUST_ORG_ID not set (optional but recommended)"
else
  echo "BRAINTRUST_ORG_ID is set: $BRAINTRUST_ORG_ID"
fi

echo ""
echo "Step 1: cleaning previous local Terraform state..."
rm -rf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup

echo "Step 2: terraform init"
terraform init -backend=false

echo ""
echo "Step 3: terraform validate"
terraform validate

echo ""
echo "Step 4: terraform plan (executes smoke-test data source)"
terraform plan

echo ""
echo "Provider smoke test completed successfully."
