# Base viewer role.
resource "braintrustdata_role" "viewer" {
  name        = "example-viewer"
  description = "Read-only access role"

  member_permissions = ["read"]
}

# Editor role that composes the viewer role and adds write capability.
resource "braintrustdata_role" "editor" {
  name        = "example-editor"
  description = "Read/write access role"

  member_permissions = ["update"]
  member_roles       = [braintrustdata_role.viewer.id]
}

output "role_ids" {
  value = {
    viewer = braintrustdata_role.viewer.id
    editor = braintrustdata_role.editor.id
  }
}
