# Read an experiment by ID.
data "braintrustdata_experiment" "by_id" {
  # replace with real ID or wire from data/resource
  id = "exp-abc123"
}

# Read an experiment by name + project context.
data "braintrustdata_experiment" "by_name" {
  name = "gpt-4-baseline"
  # replace with real ID or wire from data/resource
  project_id = "proj-abc123"
}

output "experiment_lookup" {
  value = {
    by_id = {
      id          = data.braintrustdata_experiment.by_id.id
      name        = data.braintrustdata_experiment.by_id.name
      description = data.braintrustdata_experiment.by_id.description
      public      = data.braintrustdata_experiment.by_id.public
    }
    by_name = {
      id       = data.braintrustdata_experiment.by_name.id
      created  = data.braintrustdata_experiment.by_name.created
      metadata = data.braintrustdata_experiment.by_name.metadata
      tags     = data.braintrustdata_experiment.by_name.tags
    }
  }
}
