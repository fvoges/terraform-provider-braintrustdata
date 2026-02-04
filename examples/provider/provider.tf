terraform {
  required_providers {
    braintrustdata = {
      source  = "registry.terraform.io/braintrustdata/braintrustdata"
      version = "0.1.0"
    }
  }
}

# Option 1: Configure provider with explicit values
provider "braintrustdata" {
  # api_key         = "sk-your-key-here"  # Or use BRAINTRUST_API_KEY env var
  # organization_id = "org-your-id-here"  # Or use BRAINTRUST_ORG_ID env var
  # api_url         = "https://api.braintrust.dev"  # Optional, this is the default
}

# Option 2: Use environment variables (recommended for security)
# export BRAINTRUST_API_KEY="sk-your-key-here"
# export BRAINTRUST_ORG_ID="org-your-id-here"

# Since we don't have resources yet, we'll just test provider initialization
# Once resources are implemented, you can add them here
