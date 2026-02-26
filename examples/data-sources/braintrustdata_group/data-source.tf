# Read a group by ID.
data "braintrustdata_group" "by_id" {
  # replace with real ID or wire from data/resource
  id = "group-123"
}

# Read a group by name.
data "braintrustdata_group" "by_name" {
  name = "Engineering"
}

output "group_lookup" {
  value = {
    id            = data.braintrustdata_group.by_name.id
    member_users  = data.braintrustdata_group.by_name.member_users
    member_groups = data.braintrustdata_group.by_name.member_groups
  }
}
