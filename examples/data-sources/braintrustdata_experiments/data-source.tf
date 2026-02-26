# List experiments in a project.
data "braintrustdata_experiments" "all" {
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
}

# Optional exact-name filter.
data "braintrustdata_experiments" "filtered" {
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
  name       = "gpt-4-baseline"
}

locals {
  production_experiments = [
    for exp in data.braintrustdata_experiments.all.experiments : exp
    if contains(exp.tags, "production")
  ]

  baseline_experiment = one([
    for exp in data.braintrustdata_experiments.all.experiments : exp
    if exp.name == "gpt-4-baseline"
  ])
}

output "experiment_lists" {
  value = {
    all_ids                   = data.braintrustdata_experiments.all.ids
    filtered_ids              = data.braintrustdata_experiments.filtered.ids
    production_experiment_ids = [for exp in local.production_experiments : exp.id]
    baseline_experiment_id    = local.baseline_experiment.id
  }
}
