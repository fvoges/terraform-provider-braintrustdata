# Read a role by ID
data "braintrustdata_role" "by_id" {
  id = "role-123"
}

# Read a role by name with optional organization filter
data "braintrustdata_role" "by_name" {
  name     = "admin"
  org_name = "example-org"
}

output "role_id" {
  value = data.braintrustdata_role.by_name.id
}

output "role_permissions" {
  value = data.braintrustdata_role.by_name.member_permissions
}
