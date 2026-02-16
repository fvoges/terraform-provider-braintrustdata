# List all datasets in a project
data "braintrustdata_datasets" "all" {
  project_id = "proj-abc123"
}

# Output all dataset IDs
output "all_dataset_ids" {
  value = data.braintrustdata_datasets.all.ids
}

# Output all dataset names
output "all_dataset_names" {
  value = [for ds in data.braintrustdata_datasets.all.datasets : ds.name]
}

# List datasets with a specific name filter
data "braintrustdata_datasets" "filtered" {
  project_id = "proj-abc123"
  name       = "customer-support-v1"
}

# Find a specific dataset by name
locals {
  support_datasets = [
    for ds in data.braintrustdata_datasets.all.datasets : ds
    if ds.name == "customer-support-v1"
  ]
  # The one() function ensures exactly one dataset is found
  # and provides a clearer error message if none or multiple are found
  support_dataset = one(local.support_datasets)
}

output "support_dataset_id" {
  value = local.support_dataset.id
}
