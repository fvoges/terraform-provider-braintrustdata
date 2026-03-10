# Read a function by ID.
data "braintrustdata_function" "by_id" {
  # replace with real ID or wire from data/resource
  id = "function-abc123"
}

# Read a function by project + name.
data "braintrustdata_function" "by_name" {
  project_id = "project-abc123" # replace with real project ID
  name       = "my-tool"
}

# Read a function by project + slug.
data "braintrustdata_function" "by_slug" {
  project_id = "project-abc123" # replace with real project ID
  slug       = "my-tool"
}

output "function_lookup" {
  value = {
    id            = data.braintrustdata_function.by_id.id
    name          = data.braintrustdata_function.by_id.name
    function_type = data.braintrustdata_function.by_id.function_type
    slug          = data.braintrustdata_function.by_id.slug
  }
}
