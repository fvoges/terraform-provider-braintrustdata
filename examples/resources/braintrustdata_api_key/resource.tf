resource "braintrustdata_api_key" "automation" {
  name = "terraform-automation-key"
}

output "api_key_value" {
  description = "Sensitive key material. Only available at creation time."
  value       = braintrustdata_api_key.automation.key
  sensitive   = true
}

output "api_key_preview" {
  description = "Preview string returned by the API key resource."
  value       = braintrustdata_api_key.automation.preview_name
}
