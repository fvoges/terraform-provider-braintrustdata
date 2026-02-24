locals {
  normalized_user_id_by_identity = {
    for identity, user_id in var.user_id_by_identity :
    (var.identity_case_insensitive ? lower(trimspace(identity)) : trimspace(identity)) => trimspace(user_id)
  }

  normalized_projects = {
    for project_key, project in var.projects :
    project_key => {
      name        = trimspace(coalesce(try(project.name, null), project_key))
      description = try(project.description, null)
    }
  }

  normalized_groups = {
    for group_key, group in var.groups :
    group_key => {
      name        = trimspace(coalesce(try(group.name, null), group_key))
      description = try(group.description, null)
      member_identities = [
        for identity in tolist(try(group.member_identities, [])) :
        (var.identity_case_insensitive ? lower(trimspace(identity)) : trimspace(identity))
      ]
      member_user_ids   = [for id in tolist(try(group.member_user_ids, [])) : trimspace(id)]
      member_group_keys = tolist(try(group.member_group_keys, []))
    }
  }

  missing_user_identities = toset(flatten([
    for _, group in local.normalized_groups : [
      for identity in group.member_identities : identity
      if identity != "" && !contains(keys(local.normalized_user_id_by_identity), identity)
    ]
  ]))

  missing_member_group_keys = toset(flatten([
    for _, group in local.normalized_groups : [
      for member_group_key in group.member_group_keys : member_group_key
      if !contains(keys(local.normalized_groups), member_group_key)
    ]
  ]))

  normalized_project_group_bindings = [
    for binding in var.project_group_bindings : {
      project_key          = binding.project_key
      group_key            = binding.group_key
      permissions          = sort(tolist(binding.permissions))
      restrict_object_type = try(binding.restrict_object_type, null)
    }
  ]

  missing_binding_project_keys = toset([
    for binding in local.normalized_project_group_bindings : binding.project_key
    if !contains(keys(local.normalized_projects), binding.project_key)
  ])

  missing_binding_group_keys = toset([
    for binding in local.normalized_project_group_bindings : binding.group_key
    if !contains(keys(local.normalized_groups), binding.group_key)
  ])

  group_member_user_ids = {
    for group_key, group in local.normalized_groups :
    group_key => sort(distinct(concat(
      compact(group.member_user_ids),
      [
        for identity in group.member_identities : local.normalized_user_id_by_identity[identity]
        if identity != "" && contains(keys(local.normalized_user_id_by_identity), identity)
      ]
    )))
  }

  acl_entries = flatten([
    for binding in local.normalized_project_group_bindings : [
      for permission in binding.permissions : {
        key                  = join("|", [binding.project_key, binding.group_key, permission, coalesce(binding.restrict_object_type, "*")])
        project_key          = binding.project_key
        group_key            = binding.group_key
        permission           = permission
        restrict_object_type = binding.restrict_object_type
      }
    ]
  ])

  acl_entries_by_key = {
    for entry in local.acl_entries : entry.key => entry
  }
}
