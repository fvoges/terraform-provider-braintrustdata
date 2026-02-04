# List all groups in the default organization
data "braintrustdata_groups" "all" {
  # org_id defaults to provider configuration
}

# Output all group IDs
output "all_group_ids" {
  value = data.braintrustdata_groups.all.ids
}

# Output all group names
output "all_group_names" {
  value = [for g in data.braintrustdata_groups.all.groups : g.name]
}

# Find a specific group by name using a for expression
locals {
  engineering_group = [
    for g in data.braintrustdata_groups.all.groups : g
    if g.name == "Engineering"
  ][0]
}

output "engineering_group_id" {
  value = local.engineering_group.id
}

# List groups in a specific organization
data "braintrustdata_groups" "other_org" {
  org_id = "org-456"
}
