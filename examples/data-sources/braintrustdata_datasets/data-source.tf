# List datasets in a project.
data "braintrustdata_datasets" "all" {
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
}

# Optional exact-name filter.
data "braintrustdata_datasets" "filtered" {
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
  name       = "customer-support-v1"
}

locals {
  support_datasets = [
    for ds in data.braintrustdata_datasets.all.datasets : ds
    if ds.name == "customer-support-v1"
  ]
  support_dataset = one(local.support_datasets)
}

output "dataset_lists" {
  value = {
    all_ids            = data.braintrustdata_datasets.all.ids
    filtered_ids       = data.braintrustdata_datasets.filtered.ids
    support_dataset_id = local.support_dataset.id
  }
}
