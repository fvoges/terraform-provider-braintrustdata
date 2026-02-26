# Read a project by ID
data "braintrustdata_project" "by_id" {
  id = "project-id-123"
}

# Read a project by API-native searchable attributes
data "braintrustdata_project" "by_name" {
  name     = "my-ml-project"
  org_name = "example-org"
}

output "project_id" {
  value = data.braintrustdata_project.by_name.id
}

output "project_org_id" {
  value = data.braintrustdata_project.by_name.org_id
}
