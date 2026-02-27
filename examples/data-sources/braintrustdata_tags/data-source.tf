# List tags with API-native filters
data "braintrustdata_tags" "all" {
  project_id = "proj-123"
  limit      = 50
}

# Filter tags by exact name
data "braintrustdata_tags" "filtered" {
  project_name = "example-project"
  tag_name     = "production"
}

output "all_tag_ids" {
  value = data.braintrustdata_tags.all.ids
}

output "filtered_tags" {
  value = data.braintrustdata_tags.filtered.tags
}
