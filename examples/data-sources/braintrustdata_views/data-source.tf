# List views for an object.
data "braintrustdata_views" "all" {
  # replace with real IDs or wire from data/resource
  object_id   = "proj-abc123"
  object_type = "project"
}

# Optional API-native filters.
data "braintrustdata_views" "filtered" {
  object_id   = "proj-abc123"
  object_type = "project"
  view_name   = "default"
  view_type   = "projects"
  limit       = 10
}

locals {
  default_views = [
    for view in data.braintrustdata_views.all.views : view
    if view.name == "default"
  ]
}

output "view_lists" {
  value = {
    all_ids      = data.braintrustdata_views.all.ids
    filtered_ids = data.braintrustdata_views.filtered.ids
    default_view = length(local.default_views) > 0 ? local.default_views[0].id : null
  }
}
