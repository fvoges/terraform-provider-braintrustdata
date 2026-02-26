# Read an organization by ID
data "braintrustdata_org" "by_id" {
  id = "org-123"
}

# Read an organization by exact name
data "braintrustdata_org" "by_name" {
  name = "example-org"
}

output "organization_id" {
  value = data.braintrustdata_org.by_name.id
}

output "organization_api_url" {
  value = data.braintrustdata_org.by_name.api_url
}
