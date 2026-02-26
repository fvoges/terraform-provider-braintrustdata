# List projects with no filters.
data "braintrustdata_projects" "all" {
  limit = 50
}

# Filter projects by API-native parameters.
data "braintrustdata_projects" "filtered" {
  project_name = "my-ml-project"
  org_name     = "example-org"
}

output "project_lists" {
  value = {
    all_ids   = data.braintrustdata_projects.all.ids
    filtered  = data.braintrustdata_projects.filtered.projects
    all_count = length(data.braintrustdata_projects.all.projects)
  }
}
