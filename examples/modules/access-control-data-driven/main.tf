terraform {
  required_version = ">= 1.4.0"

  required_providers {
    braintrustdata = {
      source = "braintrustdata/braintrustdata"
    }
  }
}

resource "terraform_data" "validate_inputs" {
  input = {
    project_count = length(local.normalized_projects)
    group_count   = length(local.normalized_groups)
    acl_count     = length(local.acl_entries_by_key)
  }

  lifecycle {
    precondition {
      condition     = length(local.normalized_user_identity_conflicts) == 0
      error_message = "Duplicate normalized identities in user_id_by_identity: ${join(", ", local.normalized_user_identity_conflicts)}"
    }

    precondition {
      condition     = length(local.missing_user_identities) == 0
      error_message = "Missing user_id_by_identity entries for identities: ${join(", ", sort(tolist(local.missing_user_identities)))}"
    }

    precondition {
      condition     = length(local.missing_member_group_keys) == 0
      error_message = "groups[*].member_group_keys referenced unknown groups: ${join(", ", sort(tolist(local.missing_member_group_keys)))}"
    }

    precondition {
      condition     = length(local.missing_binding_project_keys) == 0
      error_message = "project_group_bindings referenced unknown project keys: ${join(", ", sort(tolist(local.missing_binding_project_keys)))}"
    }

    precondition {
      condition     = length(local.missing_binding_group_keys) == 0
      error_message = "project_group_bindings referenced unknown group keys: ${join(", ", sort(tolist(local.missing_binding_group_keys)))}"
    }

    precondition {
      condition     = length(local.acl_duplicate_keys) == 0
      error_message = "project_group_bindings expanded to duplicate canonical ACL keys: ${join(", ", local.acl_duplicate_keys)}"
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
  for_each = local.normalized_groups

  name        = "${var.group_name_prefix}${each.value.name}"
  description = each.value.description

  member_users = local.group_member_user_ids[each.key]
  member_groups = [
    for member_group_key in sort(distinct(compact(each.value.member_group_keys))) :
    braintrustdata_group.role_groups[member_group_key].id
  ]

  depends_on = [terraform_data.validate_inputs]
}

resource "braintrustdata_acl" "project_group" {
  for_each = local.acl_entries_by_key

  object_id            = braintrustdata_project.projects[each.value.project_key].id
  object_type          = "project"
  group_id             = braintrustdata_group.role_groups[each.value.group_key].id
  permission           = each.value.permission
  restrict_object_type = each.value.restrict_object_type

  depends_on = [terraform_data.validate_inputs]
}
