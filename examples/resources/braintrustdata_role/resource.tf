# Create a basic role
resource "braintrustdata_role" "example" {
  name        = "custom-admin"
  description = "Custom administrator role with specific permissions"
}

# Role with minimal configuration
resource "braintrustdata_role" "minimal" {
  name = "viewer-role"
}

# Output role details
output "role_id" {
  value       = braintrustdata_role.example.id
  description = "The ID of the created role"
}

output "role_org_id" {
  value       = braintrustdata_role.example.org_id
  description = "The organization ID of the role"
}
