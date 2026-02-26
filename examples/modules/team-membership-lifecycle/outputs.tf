output "team_group_ids_by_key" {
  description = "Team group IDs keyed by team key."
  value = {
    for team_key, group in braintrustdata_group.team_groups :
    team_key => group.id
  }
}

output "binding_key_delimiter" {
  description = "Delimiter used when forming role_group_member_group_ids_by_binding_key."
  value       = var.binding_key_delimiter
}

output "role_group_member_group_ids_by_binding_key" {
  description = "Map of binding key project_key|role_key to team group IDs."
  value = {
    for binding_key in distinct([for pair in local.role_binding_pairs : pair.binding_key]) :
    binding_key => toset(sort(distinct([
      for pair in local.role_binding_pairs :
      braintrustdata_group.team_groups[pair.team_key].id
      if pair.binding_key == binding_key
    ])))
  }
}
