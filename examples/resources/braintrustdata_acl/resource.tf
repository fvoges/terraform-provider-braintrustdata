resource "braintrustdata_project" "example" {
  name        = "acl-example-project"
  description = "Project used for ACL examples"
}

resource "braintrustdata_group" "viewers" {
  name        = "acl-example-viewers"
  description = "Users with project read access"
}

# Group-based access: grant read on the project to the viewers group.
resource "braintrustdata_acl" "group_viewer_read" {
  object_id   = braintrustdata_project.example.id
  object_type = "project"
  group_id    = braintrustdata_group.viewers.id
  permission  = "read"
}

# User-based access: grant update for a specific user.
# replace with real ID or wire from data/resource
resource "braintrustdata_acl" "user_editor_update" {
  object_id   = braintrustdata_project.example.id
  object_type = "project"
  user_id     = "866a8a8a-fee9-4a5b-8278-12970de499c2"
  permission  = "update"
}

resource "braintrustdata_role" "restricted_viewer" {
  name        = "acl-example-restricted-viewer"
  description = "Read access restricted to experiments"
}

# Role-based access with a restricted object type scope.
resource "braintrustdata_acl" "role_restricted_read" {
  object_id            = braintrustdata_project.example.id
  object_type          = "project"
  role_id              = braintrustdata_role.restricted_viewer.id
  permission           = "read"
  restrict_object_type = "experiment"
}

output "acl_ids" {
  value = {
    group_viewer_read = braintrustdata_acl.group_viewer_read.id
    user_editor       = braintrustdata_acl.user_editor_update.id
    role_restricted   = braintrustdata_acl.role_restricted_read.id
  }
}
