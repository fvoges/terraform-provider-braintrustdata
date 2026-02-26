# Minimal project.
resource "braintrustdata_project" "minimal" {
  name = "minimal-project"
}

# Practical project used by datasets/experiments.
resource "braintrustdata_project" "evaluation" {
  name        = "customer-support-evaluation"
  description = "Project for evaluating support model quality"
}

output "project_ids" {
  value = {
    minimal    = braintrustdata_project.minimal.id
    evaluation = braintrustdata_project.evaluation.id
  }
}
