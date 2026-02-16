# Read an existing dataset by ID
data "braintrustdata_dataset" "by_id" {
  id = "ds-abc123"
}

# Read an existing dataset by name and project_id
data "braintrustdata_dataset" "by_name" {
  name       = "customer-support-v1"
  project_id = "proj-abc123"
}

# Output the dataset information
output "dataset_id" {
  value = data.braintrustdata_dataset.by_name.id
}

output "dataset_created" {
  value = data.braintrustdata_dataset.by_name.created
}

output "dataset_metadata" {
  value = data.braintrustdata_dataset.by_name.metadata
}

# Use dataset data in another resource
# For example, reference the dataset in documentation or reports
locals {
  dataset_summary = {
    id          = data.braintrustdata_dataset.by_id.id
    name        = data.braintrustdata_dataset.by_id.name
    description = data.braintrustdata_dataset.by_id.description
  }
}
