# Grant read permission to a user on a project
resource "braintrustdata_project" "example" {
  name        = "my-project"
  description = "Example project"
}

resource "braintrustdata_acl" "user_read" {
  object_id   = braintrustdata_project.example.id
  object_type = "project"
  user_id     = "user-id-here"
  permission  = "read"
}

# Grant update permission to a group on a dataset
resource "braintrustdata_group" "data_team" {
  name        = "data-team"
  description = "Data science team"
}

resource "braintrustdata_acl" "group_update" {
  object_id   = "dataset-id-here"
  object_type = "dataset"
  group_id    = braintrustdata_group.data_team.id
  permission  = "update"
}

# Grant permission to a role on an organization
resource "braintrustdata_role" "admin" {
  name        = "admin"
  description = "Administrator role"
}

resource "braintrustdata_acl" "role_org_access" {
  object_id   = "org-id-here"
  object_type = "organization"
  role_id     = braintrustdata_role.admin.id
  permission  = "create"
}

# Grant restricted access (only for specific object types)
resource "braintrustdata_acl" "restricted" {
  object_id            = braintrustdata_project.example.id
  object_type          = "project"
  user_id              = "user-id-here"
  permission           = "read"
  restrict_object_type = "experiment"
}
