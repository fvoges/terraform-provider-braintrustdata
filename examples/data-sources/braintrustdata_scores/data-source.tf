# List scores with API-native filters
data "braintrustdata_scores" "all" {
  project_id = "proj-123"
  limit      = 50
}

# Filter scores by exact name and type
data "braintrustdata_scores" "filtered" {
  project_name = "example-project"
  score_name   = "quality"
  score_type   = "categorical"
}

output "all_score_ids" {
  value = data.braintrustdata_scores.all.ids
}

output "filtered_scores" {
  value = data.braintrustdata_scores.filtered.scores
}
