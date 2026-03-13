# Create a project to own the view.
resource "braintrustdata_project" "example" {
  name = "example-view-project"
}

# Manage a view within that project.
resource "braintrustdata_view" "example" {
  object_id   = braintrustdata_project.example.id
  object_type = "project"
  view_type   = "experiments"
  name        = "example-view"

  options = jsonencode({
    freezeColumns = false
    viewType      = "table"
  })

  view_data = jsonencode({
    search = {
      filter = []
      match  = []
      sort   = []
      tag    = []
    }
  })
}

output "view_id" {
  value = braintrustdata_view.example.id
}
