output "project_ids_by_key" {
  description = "Braintrust project IDs keyed by input project key."
  value = {
    for project_key, resource in braintrustdata_project.projects :
    project_key => resource.id
  }
}

output "group_ids_by_key" {
  description = "Braintrust group IDs keyed by input group key."
  value = {
    for group_key, resource in braintrustdata_group.role_groups :
    group_key => resource.id
  }
}

output "acl_ids_by_key" {
  description = "ACL IDs keyed by canonical binding key project|group|permission|restrict_object_type."
  value = {
    for acl_key, resource in braintrustdata_acl.project_group :
    acl_key => resource.id
  }
}

output "normalized_user_identity_keys" {
  description = "The effective identity keys used for membership resolution after normalization."
  value       = sort(keys(local.normalized_user_id_by_identity))
}
