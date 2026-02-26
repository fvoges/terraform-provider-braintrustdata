# Read a project by ID.
data "braintrustdata_project" "by_id" {
  # replace with real ID or wire from data/resource
  id = "project-id-123"
}

# Read a project by searchable attributes.
data "braintrustdata_project" "by_name" {
  name     = "my-ml-project"
  org_name = "example-org"
}

output "project_lookup" {
  value = {
    by_id_id      = data.braintrustdata_project.by_id.id
    by_name_id    = data.braintrustdata_project.by_name.id
    by_name_org   = data.braintrustdata_project.by_name.org_id
    by_name_descr = data.braintrustdata_project.by_name.description
  }
}
