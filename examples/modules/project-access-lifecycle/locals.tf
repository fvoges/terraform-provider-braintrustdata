locals {
  normalized_role_catalog = {
    for role_key, role in var.role_catalog :
    trimspace(role_key) => {
      description          = try(role.description, null)
      permissions          = sort(distinct([for permission in tolist(role.permissions) : trimspace(permission)]))
      restrict_object_type = try(role.restrict_object_type, null)
    }
  }

  normalized_projects = {
    for project_key, project in var.projects :
    trimspace(project_key) => {
      name          = trimspace(coalesce(try(project.name, null), project_key))
      description   = try(project.description, null)
      enabled_roles = try(project.enabled_roles, null)
      role_overrides = {
        for role_key, override in try(project.role_overrides, {}) :
        trimspace(role_key) => {
          permissions = try(override.permissions, null) == null ? null : sort(distinct([
            for permission in tolist(override.permissions) : trimspace(permission)
          ]))
          restrict_object_type = try(override.restrict_object_type, null)
          description          = try(override.description, null)
        }
      }
    }
  }

  missing_enabled_roles = toset(flatten([
    for _, project in local.normalized_projects : [
      for role_key in coalesce(project.enabled_roles, toset(keys(local.normalized_role_catalog))) : role_key
      if !contains(keys(local.normalized_role_catalog), role_key)
    ]
  ]))

  missing_override_roles = toset(flatten([
    for _, project in local.normalized_projects : [
      for role_key, _ in project.role_overrides : role_key
      if !contains(keys(local.normalized_role_catalog), role_key)
    ]
  ]))

  effective_project_roles = {
    for project_key, project in local.normalized_projects :
    project_key => {
      for role_key in sort(tolist(coalesce(project.enabled_roles, toset(keys(local.normalized_role_catalog))))) :
      role_key => {
        description = coalesce(
          try(project.role_overrides[role_key].description, null),
          try(local.normalized_role_catalog[role_key].description, null),
          "${project_key} ${role_key} access role"
        )
        permissions = coalesce(
          try(project.role_overrides[role_key].permissions, null),
          local.normalized_role_catalog[role_key].permissions
        )
        restrict_object_type = coalesce(
          try(project.role_overrides[role_key].restrict_object_type, null),
          local.normalized_role_catalog[role_key].restrict_object_type
        )
      }
    }
  }

  role_group_bindings = flatten([
    for project_key, roles in local.effective_project_roles : [
      for role_key, role in roles : {
        binding_key = join(var.binding_key_delimiter, [project_key, role_key])
        project_key = project_key
        role_key    = role_key
        role        = role
      }
    ]
  ])

  role_group_bindings_by_key = {
    for binding in local.role_group_bindings : binding.binding_key => binding
  }

  normalized_role_group_member_group_ids_by_binding_key = {
    for binding_key, group_ids in var.role_group_member_group_ids_by_binding_key :
    trimspace(binding_key) => sort(distinct(compact([
      for group_id in tolist(group_ids) : trimspace(group_id)
    ])))
  }

  unknown_binding_keys = toset([
    for binding_key, _ in local.normalized_role_group_member_group_ids_by_binding_key : binding_key
    if !contains(keys(local.role_group_bindings_by_key), binding_key)
  ])

  acl_entries = flatten([
    for binding_key, binding in local.role_group_bindings_by_key : [
      for permission in binding.role.permissions : {
        key                  = join(var.binding_key_delimiter, [binding.project_key, binding.role_key, permission, coalesce(binding.role.restrict_object_type, "*")])
        binding_key          = binding_key
        project_key          = binding.project_key
        role_key             = binding.role_key
        permission           = permission
        restrict_object_type = binding.role.restrict_object_type
      }
    ]
  ])

  acl_entries_by_key = {
    for entry in local.acl_entries : entry.key => entry
  }

  acl_duplicate_keys = sort([
    for acl_key in distinct([for entry in local.acl_entries : entry.key]) : acl_key
    if length([
      for entry in local.acl_entries : entry.key
      if entry.key == acl_key
    ]) > 1
  ])
}
