provider "braintrustdata" {
  # Recommended: use environment variables for credentials.
  # BRAINTRUST_API_KEY
  # BRAINTRUST_ORG_ID
}

# Smoke test: performs a real API read so credentials/config are validated.
data "braintrustdata_users" "smoke" {
  limit = 1
}

output "smoke_user_ids" {
  description = "User IDs returned by the smoke-test lookup."
  value       = data.braintrustdata_users.smoke.ids
}
