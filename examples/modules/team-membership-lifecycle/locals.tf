locals {
  normalized_user_identities = [
    for identity, _ in var.user_id_by_identity :
    (var.identity_case_insensitive ? lower(trimspace(identity)) : trimspace(identity))
  ]

  normalized_user_identity_conflicts = sort([
    for normalized_identity in distinct(local.normalized_user_identities) : normalized_identity
    if normalized_identity != "" && length([
      for candidate in local.normalized_user_identities : candidate
      if candidate == normalized_identity
    ]) > 1
  ])

  normalized_user_id_by_identity = {
    for identity, user_id in var.user_id_by_identity :
    (var.identity_case_insensitive ? lower(trimspace(identity)) : trimspace(identity)) => trimspace(user_id)
  }

  normalized_teams = {
    for team_key, team in var.teams :
    trimspace(team_key) => {
      name        = trimspace(coalesce(try(team.name, null), team_key))
      description = try(team.description, null)
      member_identities = sort(distinct(compact([
        for identity in tolist(try(team.member_identities, [])) :
        (var.identity_case_insensitive ? lower(trimspace(identity)) : trimspace(identity))
      ])))
      member_user_ids = sort(distinct(compact([
        for user_id in tolist(try(team.member_user_ids, [])) : trimspace(user_id)
      ])))
      member_team_keys = sort(distinct(compact([
        for member_team_key in tolist(try(team.member_team_keys, [])) : trimspace(member_team_key)
      ])))
      project_role_bindings = values({
        for binding in tolist(try(team.project_role_bindings, [])) :
        "${trimspace(binding.project_key)}|${trimspace(binding.role_key)}" => {
          project_key = trimspace(binding.project_key)
          role_key    = trimspace(binding.role_key)
        }
        if trimspace(binding.project_key) != "" && trimspace(binding.role_key) != ""
      })
    }
  }

  missing_member_team_keys = toset(flatten([
    for _, team in local.normalized_teams : [
      for member_team_key in team.member_team_keys : member_team_key
      if !contains(keys(local.normalized_teams), member_team_key)
    ]
  ]))

  unresolved_identities = sort(distinct(flatten([
    for _, team in local.normalized_teams : [
      for identity in team.member_identities : identity
      if !contains(keys(local.normalized_user_id_by_identity), identity)
    ]
  ])))

  identities_to_lookup = var.resolve_identities_with_data_source ? toset(local.unresolved_identities) : toset([])

  role_binding_pairs = values({
    for pair in flatten([
      for team_key, team in local.normalized_teams : [
        for binding in team.project_role_bindings : {
          team_key    = team_key
          project_key = binding.project_key
          role_key    = binding.role_key
          binding_key = "${binding.project_key}|${binding.role_key}"
        }
      ]
    ]) :
    "${pair.team_key}|${pair.binding_key}" => pair
  })
}
