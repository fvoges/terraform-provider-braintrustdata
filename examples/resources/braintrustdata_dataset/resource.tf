resource "braintrustdata_project" "evaluation" {
  name        = "dataset-example-project"
  description = "Project for dataset examples"
}

# Minimal dataset.
resource "braintrustdata_dataset" "minimal" {
  name       = "customer-support-v1"
  project_id = braintrustdata_project.evaluation.id
}

# Practical dataset with metadata for downstream workflows.
resource "braintrustdata_dataset" "curated" {
  name        = "customer-support-curated"
  project_id  = braintrustdata_project.evaluation.id
  description = "Curated evaluation dataset"

  metadata = {
    version      = "1.0"
    source       = "production-logs"
    sample_count = "10000"
    owner        = "evaluation-platform"
  }
}

output "dataset_ids" {
  value = {
    minimal = braintrustdata_dataset.minimal.id
    curated = braintrustdata_dataset.curated.id
  }
}
