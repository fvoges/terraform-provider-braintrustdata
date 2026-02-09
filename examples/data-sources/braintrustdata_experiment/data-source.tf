# Read an existing experiment by ID
data "braintrustdata_experiment" "by_id" {
  id = "exp-abc123"
}

# Read an existing experiment by name and project_id
data "braintrustdata_experiment" "by_name" {
  name       = "gpt-4-baseline"
  project_id = "proj-abc123"
}

# Output the experiment information
output "experiment_id" {
  value = data.braintrustdata_experiment.by_name.id
}

output "experiment_created" {
  value = data.braintrustdata_experiment.by_name.created
}

output "experiment_metadata" {
  value = data.braintrustdata_experiment.by_name.metadata
}

output "experiment_tags" {
  value = data.braintrustdata_experiment.by_name.tags
}

# Use experiment data in another resource
# For example, reference the experiment in documentation or reports
locals {
  experiment_summary = {
    id          = data.braintrustdata_experiment.by_id.id
    name        = data.braintrustdata_experiment.by_id.name
    description = data.braintrustdata_experiment.by_id.description
    is_public   = data.braintrustdata_experiment.by_id.public
  }
}
