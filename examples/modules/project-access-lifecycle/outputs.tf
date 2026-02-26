output "project_ids_by_key" {
  description = "Project IDs keyed by input project key."
  value = {
    for project_key, project in braintrustdata_project.projects :
    project_key => project.id
  }
}

output "role_group_ids_by_binding_key" {
  description = "Role-group IDs keyed by binding key project_key|role_key."
  value = {
    for binding_key, group in braintrustdata_group.role_groups :
    binding_key => group.id
  }
}

output "role_group_names_by_binding_key" {
  description = "Role-group names keyed by binding key project_key|role_key."
  value = {
    for binding_key, group in braintrustdata_group.role_groups :
    binding_key => group.name
  }
}
