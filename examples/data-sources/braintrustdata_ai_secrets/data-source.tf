# List AI secrets with API-native filters
data "braintrustdata_ai_secrets" "all" {
  limit = 50
}

# Filter AI secrets by exact name, type, and IDs
data "braintrustdata_ai_secrets" "filtered" {
  ai_secret_name  = "OPENAI_API_KEY"
  ai_secret_types = ["openai"]
  filter_ids      = ["ai-secret-123"]
  org_name        = "example-org"
}

output "all_ai_secret_ids" {
  value = data.braintrustdata_ai_secrets.all.ids
}

output "filtered_ai_secrets" {
  value = data.braintrustdata_ai_secrets.filtered.ai_secrets
}
