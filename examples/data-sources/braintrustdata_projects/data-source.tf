# List projects with no filters
data "braintrustdata_projects" "all" {
  limit = 50
}

# Filter projects by API-native project_name
data "braintrustdata_projects" "filtered" {
  project_name = "my-ml-project"
  org_name     = "example-org"
}

output "all_project_ids" {
  value = data.braintrustdata_projects.all.ids
}

output "filtered_projects" {
  value = data.braintrustdata_projects.filtered.projects
}
