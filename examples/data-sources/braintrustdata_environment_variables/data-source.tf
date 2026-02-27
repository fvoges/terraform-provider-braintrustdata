# List environment variables for a project
data "braintrustdata_environment_variables" "all" {
  object_type = "project"
  object_id   = "project-123"
}

# Optional exact-name filter after retrieval
data "braintrustdata_environment_variables" "filtered" {
  object_type = "project"
  object_id   = "project-123"
  name        = "OPENAI_API_KEY"
}

output "environment_variable_ids" {
  value = data.braintrustdata_environment_variables.all.ids
}

output "filtered_environment_variables" {
  value = data.braintrustdata_environment_variables.filtered.environment_variables
}
