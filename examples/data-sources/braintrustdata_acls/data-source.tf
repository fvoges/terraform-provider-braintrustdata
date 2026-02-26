# List ACLs for a specific object using API-native filters
data "braintrustdata_acls" "all" {
  object_id   = "project-123"
  object_type = "project"
  limit       = 50
}

output "acl_ids" {
  value = data.braintrustdata_acls.all.ids
}

output "acls" {
  value = data.braintrustdata_acls.all.acls
}
