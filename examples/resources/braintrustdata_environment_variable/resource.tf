resource "braintrustdata_project" "evaluation" {
  name        = "env-var-example-project"
  description = "Project used by environment variable examples"
}

# Managed environment variable on a project.
resource "braintrustdata_environment_variable" "openai_api_key" {
  object_type = "project"
  object_id   = braintrustdata_project.evaluation.id
  name        = "OPENAI_API_KEY"
  value       = "sk-live-replace-me"

  metadata = {
    owner   = "ml-platform"
    purpose = "evaluation-runtime"
  }
}

output "environment_variable_id" {
  value = braintrustdata_environment_variable.openai_api_key.id
}
