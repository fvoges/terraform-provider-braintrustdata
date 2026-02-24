# Read a user by ID
data "braintrustdata_user" "by_id" {
  id = "866a8a8a-fee9-4a5b-8278-12970de499c2"
}

# Read a user by API-native searchable attributes
data "braintrustdata_user" "by_email" {
  email    = "alice@example.com"
  org_name = "example-org"
}

output "user_id" {
  value = data.braintrustdata_user.by_email.id
}

output "user_email" {
  value = data.braintrustdata_user.by_email.email
}
