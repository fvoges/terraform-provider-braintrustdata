terraform {
  required_version = ">= 1.4.0"

  required_providers {
    braintrustdata = {
      source  = "braintrustdata/braintrustdata"
      version = "= 0.1.0"
    }
  }
}

resource "terraform_data" "validate_inputs" {
  input = {
    project_count   = length(local.normalized_projects)
    role_count      = length(local.normalized_role_catalog)
    binding_count   = length(local.role_group_bindings_by_key)
    acl_key_count   = length(local.acl_entries_by_key)
    acl_entry_count = length(local.acl_entries)
  }

  lifecycle {
    precondition {
      condition     = length(local.missing_enabled_roles) == 0
      error_message = "projects[*].enabled_roles referenced unknown roles: ${join(", ", sort(tolist(local.missing_enabled_roles)))}"
    }

    precondition {
      condition     = length(local.missing_override_roles) == 0
      error_message = "projects[*].role_overrides referenced unknown roles: ${join(", ", sort(tolist(local.missing_override_roles)))}"
    }

    precondition {
      condition     = length(local.unknown_binding_keys) == 0
      error_message = "role_group_member_group_ids_by_binding_key contains unknown binding keys: ${join(", ", sort(tolist(local.unknown_binding_keys)))}"
    }

    precondition {
      condition     = length(local.acl_duplicate_keys) == 0
      error_message = "effective ACL expansion generated duplicate keys: ${join(", ", local.acl_duplicate_keys)}"
    }
  }
}

resource "braintrustdata_project" "projects" {
  for_each = local.normalized_projects

  name        = "${var.project_name_prefix}${each.value.name}"
  description = each.value.description

  depends_on = [terraform_data.validate_inputs]
}

resource "braintrustdata_group" "role_groups" {
  for_each = local.role_group_bindings_by_key

  name        = "${var.role_group_name_prefix}${each.value.project_key}-${each.value.role_key}"
  description = each.value.role.description

  member_users = []
  member_groups = try(
    local.normalized_role_group_member_group_ids_by_binding_key[each.key],
    []
  )

  depends_on = [terraform_data.validate_inputs]
}

resource "braintrustdata_acl" "project_role_access" {
  for_each = local.acl_entries_by_key

  object_id            = braintrustdata_project.projects[each.value.project_key].id
  object_type          = "project"
  group_id             = braintrustdata_group.role_groups[each.value.binding_key].id
  permission           = each.value.permission
  restrict_object_type = each.value.restrict_object_type

  depends_on = [terraform_data.validate_inputs]
}
