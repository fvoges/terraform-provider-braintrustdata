# Minimal group.
resource "braintrustdata_group" "support_reviewers" {
  name        = "support-reviewers"
  description = "Users who review support experiments"
}

# Group with direct users plus nested group membership.
resource "braintrustdata_group" "ml_team" {
  name        = "ml-team"
  description = "Machine learning team"

  # replace with real ID or wire from data/resource
  member_users  = ["866a8a8a-fee9-4a5b-8278-12970de499c2"]
  member_groups = [braintrustdata_group.support_reviewers.id]
}

# Optional org-scoped group.
resource "braintrustdata_group" "org_admins" {
  name        = "org-admins"
  description = "Organization administrators"

  # replace with real ID or wire from data/resource
  org_id = "org-specific-id"
}

output "group_ids" {
  value = {
    support_reviewers = braintrustdata_group.support_reviewers.id
    ml_team           = braintrustdata_group.ml_team.id
    org_admins        = braintrustdata_group.org_admins.id
  }
}
