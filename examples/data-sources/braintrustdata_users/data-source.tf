# List users with no filters.
data "braintrustdata_users" "all" {
  limit = 50
}

# Filter users by email.
data "braintrustdata_users" "filtered_email" {
  email = "alice@example.com"
}

# Filter users by IDs using API-native repeated ids query parameter.
data "braintrustdata_users" "filtered_ids" {
  filter_ids = data.braintrustdata_users.all.ids
}

output "user_lists" {
  value = {
    all_ids            = data.braintrustdata_users.all.ids
    filtered_email_ids = [for user in data.braintrustdata_users.filtered_email.users : user.id]
    filtered_ids_count = length(data.braintrustdata_users.filtered_ids.users)
  }
}
