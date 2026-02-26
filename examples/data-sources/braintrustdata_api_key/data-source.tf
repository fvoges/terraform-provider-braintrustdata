# Read an API key by ID
data "braintrustdata_api_key" "by_id" {
  id = "api-key-123"
}

# Read an API key by name with optional organization filter
data "braintrustdata_api_key" "by_name" {
  name     = "service-key"
  org_name = "example-org"
}

output "api_key_id" {
  value = data.braintrustdata_api_key.by_name.id
}

output "api_key_preview" {
  value = data.braintrustdata_api_key.by_name.preview_name
}
