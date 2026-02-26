# Read a user by ID.
data "braintrustdata_user" "by_id" {
  # replace with real ID or wire from data/resource
  id = "866a8a8a-fee9-4a5b-8278-12970de499c2"
}

# Read a user by searchable attributes.
data "braintrustdata_user" "by_email" {
  email    = "alice@example.com"
  org_name = "example-org"
}

output "user_lookup" {
  value = {
    by_id_id      = data.braintrustdata_user.by_id.id
    by_email_id   = data.braintrustdata_user.by_email.id
    by_email_mail = data.braintrustdata_user.by_email.email
  }
}
