resource "braintrustdata_project" "evaluation" {
  name        = "experiment-example-project"
  description = "Project for experiment examples"
}

# Minimal experiment.
resource "braintrustdata_experiment" "minimal" {
  name       = "gpt-4-baseline"
  project_id = braintrustdata_project.evaluation.id
}

# Practical experiment with metadata, tags, and repository provenance.
resource "braintrustdata_experiment" "production_candidate" {
  name        = "prompt-optimization-v1"
  project_id  = braintrustdata_project.evaluation.id
  description = "Candidate prompt tuned for support responses"
  public      = false

  metadata = {
    model       = "gpt-4"
    temperature = "0.7"
    dataset     = "customer-support-curated"
  }

  tags = ["customer-support", "production-candidate"]

  repo_info = {
    # replace with real ID or wire from data/resource
    commit         = "abc123def456"
    branch         = "main"
    tag            = "v1.0.0"
    dirty          = false
    author_name    = "Jane Developer"
    author_email   = "jane@example.com"
    commit_message = "Tune support prompt"
    commit_time    = "2026-02-18T12:00:00Z"
  }
}

output "experiment_ids" {
  value = {
    minimal              = braintrustdata_experiment.minimal.id
    production_candidate = braintrustdata_experiment.production_candidate.id
  }
}
