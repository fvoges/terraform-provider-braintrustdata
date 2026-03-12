resource "braintrustdata_project" "evaluation" {
  name        = "score-example-project"
  description = "Project for score resource examples"
}

resource "braintrustdata_score" "quality" {
  project_id  = braintrustdata_project.evaluation.id
  name        = "quality"
  score_type  = "free-form"
  description = "Free-form score for evaluator feedback"
}

output "score_id" {
  value = braintrustdata_score.quality.id
}
