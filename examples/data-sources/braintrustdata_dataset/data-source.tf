# Read a dataset by ID.
data "braintrustdata_dataset" "by_id" {
  # replace with real ID or wire from data/resource
  id = "ds-abc123"
}

# Read a dataset by name + project context.
data "braintrustdata_dataset" "by_name" {
  name = "customer-support-v1"
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
}

output "dataset_lookup" {
  value = {
    by_id = {
      id          = data.braintrustdata_dataset.by_id.id
      name        = data.braintrustdata_dataset.by_id.name
      description = data.braintrustdata_dataset.by_id.description
    }
    by_name = {
      id       = data.braintrustdata_dataset.by_name.id
      created  = data.braintrustdata_dataset.by_name.created
      metadata = data.braintrustdata_dataset.by_name.metadata
    }
  }
}
