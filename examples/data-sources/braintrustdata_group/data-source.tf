# Read an existing group by ID
data "braintrustdata_group" "by_id" {
  id = "group-123"
}

# Read an existing group by name
data "braintrustdata_group" "by_name" {
  name = "Engineering"
}

# Output the group information
output "group_id" {
  value = data.braintrustdata_group.by_name.id
}

output "group_members" {
  value = data.braintrustdata_group.by_name.member_ids
}

# Use group data in another resource (example)
# resource "braintrustdata_acl" "example" {
#   group_id = data.braintrustdata_group.by_name.id
#   # ... other configuration
# }
