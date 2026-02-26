# List organizations with optional filters.
data "braintrustdata_orgs" "all" {}

data "braintrustdata_orgs" "filtered" {
  org_name = "example-org"
  limit    = 10
}

output "organization_ids" {
  value = data.braintrustdata_orgs.filtered.ids
}
