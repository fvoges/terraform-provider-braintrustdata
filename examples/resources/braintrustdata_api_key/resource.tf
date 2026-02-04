# Create a basic API key
resource "braintrustdata_api_key" "example" {
  name = "my-api-key"
}

# Access the key value (only available at creation time)
output "api_key_value" {
  value     = braintrustdata_api_key.example.key
  sensitive = true
}

# The preview name shows a shortened version of the key
output "api_key_preview" {
  value = braintrustdata_api_key.example.preview_name
}
