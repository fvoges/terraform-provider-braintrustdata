# Read an AI secret by ID
data "braintrustdata_ai_secret" "by_id" {
  id = "ai-secret-123"
}

# Read an AI secret by name with optional organization and type filters
data "braintrustdata_ai_secret" "by_name" {
  name           = "OPENAI_API_KEY"
  org_name       = "example-org"
  ai_secret_type = "openai"
}

output "ai_secret_id" {
  value = data.braintrustdata_ai_secret.by_name.id
}

output "ai_secret_preview" {
  value = data.braintrustdata_ai_secret.by_name.preview_secret
}
