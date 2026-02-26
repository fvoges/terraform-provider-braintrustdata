# List API keys with API-native filters
data "braintrustdata_api_keys" "all" {
  limit = 50
}

# Filter API keys by exact name
data "braintrustdata_api_keys" "filtered" {
  api_key_name = "service-key"
  org_name     = "example-org"
}

output "all_api_key_ids" {
  value = data.braintrustdata_api_keys.all.ids
}

output "filtered_api_keys" {
  value = data.braintrustdata_api_keys.filtered.api_keys
}
