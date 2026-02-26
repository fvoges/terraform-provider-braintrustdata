# Read an environment variable by ID
data "braintrustdata_environment_variable" "by_id" {
  id = "env-var-123"
}

# Read an environment variable by name and owner object
data "braintrustdata_environment_variable" "by_name" {
  name        = "OPENAI_API_KEY"
  object_type = "project"
  object_id   = "project-123"
}

output "environment_variable_id" {
  value = data.braintrustdata_environment_variable.by_name.id
}

output "environment_variable_created" {
  value = data.braintrustdata_environment_variable.by_name.created
}
