# List users with no filters
data "braintrustdata_users" "all" {
  limit = 50
}

# Filter users by email
data "braintrustdata_users" "filtered_email" {
  email = "alice@example.com"
}

# Filter users by IDs using API-native ids query parameter
data "braintrustdata_users" "filtered_ids" {
  filter_ids = data.braintrustdata_users.all.ids
}

output "all_user_ids" {
  value = data.braintrustdata_users.all.ids
}

output "filtered_users" {
  value = data.braintrustdata_users.filtered_email.users
}
