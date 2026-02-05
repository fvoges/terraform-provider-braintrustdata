# List all experiments in a project
data "braintrustdata_experiments" "all" {
  project_id = "proj-abc123"
}

# Output all experiment IDs
output "all_experiment_ids" {
  value = data.braintrustdata_experiments.all.ids
}

# Output all experiment names
output "all_experiment_names" {
  value = [for exp in data.braintrustdata_experiments.all.experiments : exp.name]
}

# List experiments with a specific name filter
data "braintrustdata_experiments" "filtered" {
  project_id = "proj-abc123"
  name       = "gpt-4-baseline"
}

# Find experiments by tag using for expression
locals {
  production_experiments = [
    for exp in data.braintrustdata_experiments.all.experiments : exp
    if contains(exp.tags, "production")
  ]
}

output "production_experiment_ids" {
  value = [for exp in local.production_experiments : exp.id]
}

# Find a specific experiment by name
locals {
  baseline_experiments = [
    for exp in data.braintrustdata_experiments.all.experiments : exp
    if exp.name == "gpt-4-baseline"
  ]
  # The one() function ensures exactly one experiment is found
  # and provides a clearer error message if none or multiple are found
  baseline_experiment = one(local.baseline_experiments)
}

output "baseline_experiment_id" {
  value = local.baseline_experiment.id
}

# Count experiments by public/private status
locals {
  public_count  = length([for exp in data.braintrustdata_experiments.all.experiments : exp if exp.public])
  private_count = length([for exp in data.braintrustdata_experiments.all.experiments : exp if !exp.public])
}

output "experiment_stats" {
  value = {
    total   = length(data.braintrustdata_experiments.all.experiments)
    public  = local.public_count
    private = local.private_count
  }
}
