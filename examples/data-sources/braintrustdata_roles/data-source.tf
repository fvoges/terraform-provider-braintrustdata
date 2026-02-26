# List roles with API-native filters
data "braintrustdata_roles" "all" {
  limit = 50
}

# Filter roles by exact name
data "braintrustdata_roles" "filtered" {
  role_name = "admin"
  org_name  = "example-org"
}

output "all_role_ids" {
  value = data.braintrustdata_roles.all.ids
}

output "filtered_roles" {
  value = data.braintrustdata_roles.filtered.roles
}
