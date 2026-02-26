# List groups in the provider organization.
data "braintrustdata_groups" "all" {}

# List groups in another organization.
data "braintrustdata_groups" "other_org" {
  # replace with real ID or wire from data/resource
  org_id = "org-456"
}

locals {
  engineering_group = [
    for g in data.braintrustdata_groups.all.groups : g
    if g.name == "Engineering"
  ][0]
}

output "group_lists" {
  value = {
    all_ids              = data.braintrustdata_groups.all.ids
    all_names            = [for g in data.braintrustdata_groups.all.groups : g.name]
    engineering_group_id = local.engineering_group.id
  }
}
