#!/bin/bash
set -e

echo "=========================================="
echo "Testing Braintrust Terraform Provider"
echo "=========================================="
echo ""

# Check if API key is set
if [ -z "$BRAINTRUST_API_KEY" ]; then
  echo "❌ BRAINTRUST_API_KEY environment variable not set"
  echo ""
  echo "Please set your API key:"
  echo "  export BRAINTRUST_API_KEY=\"sk-your-key-here\""
  echo ""
  echo "Get your API key from: https://www.braintrust.dev"
  exit 1
fi

echo "✅ BRAINTRUST_API_KEY is set"

# Check if org ID is set
if [ -z "$BRAINTRUST_ORG_ID" ]; then
  echo "⚠️  BRAINTRUST_ORG_ID environment variable not set (optional)"
else
  echo "✅ BRAINTRUST_ORG_ID is set: $BRAINTRUST_ORG_ID"
fi

echo ""
echo "Step 1: Cleaning previous state..."
rm -rf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup

echo "Step 2: Initializing Terraform..."
terraform init

echo ""
echo "Step 3: Validating configuration..."
terraform validate

echo ""
echo "Step 4: Testing provider (terraform plan)..."
terraform plan

echo ""
echo "=========================================="
echo "✅ Provider test completed successfully!"
echo "=========================================="
echo ""
echo "The provider can:"
echo "  ✅ Initialize with your API key"
echo "  ✅ Parse configuration correctly"
echo "  ✅ Validate without errors"
echo ""
echo "Next steps:"
echo "  - Add resources (projects, experiments, etc.)"
echo "  - Test resource creation with 'terraform apply'"
