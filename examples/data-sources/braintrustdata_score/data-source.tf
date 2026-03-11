# Read a score by ID
data "braintrustdata_score" "by_id" {
  id = "score-123"
}

# Read a score by searchable attributes
data "braintrustdata_score" "by_name" {
  name       = "quality"
  project_id = "proj-123"
  score_type = "categorical"
}

output "score_id" {
  value = data.braintrustdata_score.by_name.id
}

output "score_config_json" {
  value = data.braintrustdata_score.by_name.config
}
