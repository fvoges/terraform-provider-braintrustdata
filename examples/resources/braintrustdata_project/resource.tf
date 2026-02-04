# Create a basic project
resource "braintrustdata_project" "example" {
  name        = "my-ml-project"
  description = "ML evaluation project for tracking model performance"
}

# Project with minimal configuration
resource "braintrustdata_project" "minimal" {
  name = "minimal-project"
}

# Output project details
output "project_id" {
  value       = braintrustdata_project.example.id
  description = "The ID of the created project"
}

output "project_org_id" {
  value       = braintrustdata_project.example.org_id
  description = "The organization ID of the project"
}
