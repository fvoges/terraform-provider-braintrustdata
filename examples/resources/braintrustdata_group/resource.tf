terraform {
  required_providers {
    braintrustdata = {
      source = "braintrustdata/braintrustdata"
    }
  }
}

# Create a simple group
resource "braintrustdata_group" "ml_team" {
  name        = "ml-team"
  description = "Machine Learning Team"
}

# Create a group with members
resource "braintrustdata_group" "reviewers" {
  name        = "experiment-reviewers"
  description = "Users who can review experiments"
  member_ids  = ["user-123", "user-456"]
}

# Create a group in a specific organization
resource "braintrustdata_group" "org_admins" {
  name        = "org-admins"
  description = "Organization administrators"
  org_id      = "org-specific-id"
}
