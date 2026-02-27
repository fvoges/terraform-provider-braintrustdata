# Read a view by ID.
data "braintrustdata_view" "by_id" {
  # replace with real IDs or wire from data/resource
  id          = "view-abc123"
  object_id   = "proj-abc123"
  object_type = "project"
}

# Read a view by searchable attributes.
data "braintrustdata_view" "by_name" {
  name        = "default"
  object_id   = "proj-abc123"
  object_type = "project"
  view_type   = "projects"
}

output "view_lookup" {
  value = {
    by_id = {
      id        = data.braintrustdata_view.by_id.id
      name      = data.braintrustdata_view.by_id.name
      view_type = data.braintrustdata_view.by_id.view_type
    }
    by_name = {
      id          = data.braintrustdata_view.by_name.id
      object_id   = data.braintrustdata_view.by_name.object_id
      object_type = data.braintrustdata_view.by_name.object_type
      created     = data.braintrustdata_view.by_name.created
    }
  }
}
