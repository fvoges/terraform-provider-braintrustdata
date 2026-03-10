# List functions filtered by project + name.
data "braintrustdata_functions" "by_name" {
  project_id = "project-abc123" # replace with real project ID
  name       = "my-tool"
  limit      = 10
}

# List functions filtered by project + slug.
data "braintrustdata_functions" "by_slug" {
  project_id = "project-abc123" # replace with real project ID
  slug       = "my-tool"
  limit      = 10
}

# List functions using cursor-based pagination.
data "braintrustdata_functions" "paged" {
  starting_after = "function-abc123"
  limit          = 10
}

output "function_lists" {
  value = {
    by_name_ids = data.braintrustdata_functions.by_name.ids
    by_slug_ids = data.braintrustdata_functions.by_slug.ids
    paged_ids   = data.braintrustdata_functions.paged.ids
  }
}
