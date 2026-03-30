resource "braintrustdata_ai_secret" "openai" {
  name   = "PROVIDER_OPENAI_CREDENTIAL"
  type   = "openai"
  secret = "replace-me-with-a-real-secret"

  metadata = {
    owner = "ml-platform"
  }
}

output "ai_secret_id" {
  value = braintrustdata_ai_secret.openai.id
}

output "ai_secret_preview" {
  value = braintrustdata_ai_secret.openai.preview_secret
}
