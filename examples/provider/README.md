# Testing the Provider Locally

This guide shows how to test the Braintrust Terraform provider on your local machine.

## Prerequisites

1. **Braintrust API Key**: Get one from [braintrust.dev](https://www.braintrust.dev)
2. **Terraform**: Version 1.0 or later
3. **Built Provider**: Run `make install` from the project root

## Quick Test

### Option 1: Automated Test Script

```bash
# Set your credentials
export BRAINTRUST_API_KEY="sk-your-key-here"
export BRAINTRUST_ORG_ID="org-your-id-here"  # Optional

# Run the test script
./test-provider.sh
```

### Option 2: Manual Testing

```bash
# 1. Set credentials
export BRAINTRUST_API_KEY="sk-your-key-here"
export BRAINTRUST_ORG_ID="org-your-id-here"

# 2. Initialize Terraform
terraform init

# 3. Validate configuration
terraform validate

# 4. Test provider initialization
terraform plan
```

## What Gets Tested

Currently, the provider validates:

✅ **Provider Initialization**
- Reads API key from environment or config
- Reads organization ID from environment or config
- Validates required fields
- Creates API client with proper configuration

✅ **Security Features**
- HTTPS-only enforcement
- TLS 1.2+ configuration
- Sensitive data masking
- Bearer token authentication

✅ **Configuration Precedence**
- Environment variables work
- Provider block configuration works
- Correct precedence order (config > env)

## Expected Output

### Successful Test

```
✅ BRAINTRUST_API_KEY is set
✅ BRAINTRUST_ORG_ID is set: org-123

Step 1: Cleaning previous state...
Step 2: Initializing Terraform...

Terraform has been successfully initialized!

Step 3: Validating configuration...
Success! The configuration is valid.

Step 4: Testing provider (terraform plan)...
No changes. Your infrastructure matches the configuration.

✅ Provider test completed successfully!
```

### Common Errors and Solutions

#### Error: Missing API Key

```
Error: Missing API Key Configuration
```

**Solution**: Set the `BRAINTRUST_API_KEY` environment variable:
```bash
export BRAINTRUST_API_KEY="sk-your-key-here"
```

#### Error: HTTP URLs Not Allowed

```
panic: http:// URLs are not allowed, must use https:// for security
```

**Solution**: Ensure `api_url` uses `https://` or omit it to use the default.

#### Error: Invalid API Key

```
Error: API error: status 401, message: Unauthorized
```

**Solution**: Check that your API key is valid and hasn't been revoked.

## Testing Different Configurations

### Test 1: Environment Variables Only

```bash
export BRAINTRUST_API_KEY="sk-test"
export BRAINTRUST_ORG_ID="org-test"
terraform init
terraform validate
```

### Test 2: Provider Block Configuration

Edit `provider.tf`:
```terraform
provider "braintrustdata" {
  api_key         = "sk-test"
  organization_id = "org-test"
}
```

```bash
terraform init
terraform validate
```

### Test 3: Custom API URL

```terraform
provider "braintrustdata" {
  api_key = "sk-test"
  api_url = "https://api.staging.braintrust.dev"
}
```

### Test 4: Configuration Precedence

```bash
# Set environment variable
export BRAINTRUST_API_KEY="sk-env-key"

# Provider config will override
# provider.tf: api_key = "sk-config-key"

terraform init
# Provider should use "sk-config-key" (config wins)
```

## Advanced Testing

### Test Provider Schema

```bash
cd examples/provider
terraform providers schema -json | jq '.provider_schemas["registry.terraform.io/braintrustdata/braintrustdata"]'
```

Expected output shows:
- `api_key` (string, optional, sensitive)
- `api_url` (string, optional)
- `organization_id` (string, optional)

### Test HTTPS Enforcement

Try to set an http:// URL:

```terraform
provider "braintrustdata" {
  api_key = "sk-test"
  api_url = "http://api.braintrust.dev"  # Should fail
}
```

```bash
terraform init
# Should panic with: "http:// URLs are not allowed"
```

### Verify TLS Configuration

Create a test that verifies TLS 1.2+ is enforced:

```bash
# The provider's tests cover this
cd ../..
go test ./internal/client/... -run TestTLSConfiguration -v
```

## Debugging

### Enable Debug Logging

```bash
export TF_LOG=DEBUG
terraform init
```

### Check Provider Installation

```bash
ls -la ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/
```

Should show the provider binary for your platform.

### Verify Provider Version

```bash
terraform version
terraform providers
```

## Next Steps

Once resources are implemented (projects, datasets, experiments), you can test:

1. **Resource Creation**: `terraform apply`
2. **Resource Import**: `terraform import`
3. **Resource Updates**: Modify config and `terraform apply`
4. **Resource Deletion**: `terraform destroy`

## Cleanup

```bash
# Remove test state
rm -rf .terraform .terraform.lock.hcl terraform.tfstate*

# Uninstall provider (optional)
rm -rf ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/
```

## Troubleshooting

### Provider Not Found

If Terraform can't find the provider:

```bash
# Reinstall
cd ../..
make install

# Verify installation
ls ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/darwin_arm64/
```

### Architecture Mismatch

The Makefile assumes `darwin_arm64`. For other platforms:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o terraform-provider-braintrustdata
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/linux_amd64
cp terraform-provider-braintrustdata ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/linux_amd64/

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o terraform-provider-braintrustdata
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/darwin_amd64
cp terraform-provider-braintrustdata ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/darwin_amd64/
```

## Testing Checklist

- [ ] Provider builds successfully
- [ ] Provider installs to local plugin directory
- [ ] `terraform init` succeeds
- [ ] `terraform validate` succeeds
- [ ] API key validation works (error when missing)
- [ ] Environment variables are read correctly
- [ ] Provider configuration overrides environment
- [ ] HTTPS-only enforcement works
- [ ] Custom API URL works
- [ ] Organization ID is passed to client
