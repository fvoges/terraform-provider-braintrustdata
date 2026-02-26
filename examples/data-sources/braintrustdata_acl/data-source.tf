# Read a single ACL by ID
data "braintrustdata_acl" "example" {
  id = "acl-123"
}

output "acl_object_id" {
  value = data.braintrustdata_acl.example.object_id
}

output "acl_permission" {
  value = data.braintrustdata_acl.example.permission
}
