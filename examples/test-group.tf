terraform {
  required_providers {
    braintrustdata = {
      source  = "registry.terraform.io/braintrustdata/braintrustdata"
      version = "0.1.0"
    }
  }
}

provider "braintrustdata" {
  # Credentials from environment:
  # export BRAINTRUST_API_KEY="sk-your-key"
  # export BRAINTRUST_ORG_ID="org-your-id"
}

# Simple test group - UPDATED
resource "braintrustdata_group" "test" {
  name        = "terraform-test-group-updated"
  description = "Testing Terraform provider - UPDATED"
}

# Output the group ID
output "group_id" {
  value = braintrustdata_group.test.id
}

output "group_name" {
  value = braintrustdata_group.test.name
}

output "created_at" {
  value = braintrustdata_group.test.created
}
