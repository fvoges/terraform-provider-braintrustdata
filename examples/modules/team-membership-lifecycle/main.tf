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
    team_count       = length(local.normalized_teams)
    unresolved_count = length(local.unresolved_identities)
  }

  lifecycle {
    precondition {
      condition     = length(local.normalized_user_identity_conflicts) == 0
      error_message = "Duplicate normalized identities in user_id_by_identity: ${join(", ", local.normalized_user_identity_conflicts)}"
    }

    precondition {
      condition     = length(local.missing_member_team_keys) == 0
      error_message = "teams[*].member_team_keys referenced unknown teams: ${join(", ", sort(tolist(local.missing_member_team_keys)))}"
    }

    precondition {
      condition     = var.resolve_identities_with_data_source || length(local.unresolved_identities) == 0
      error_message = "Missing user_id_by_identity entries for identities and resolve_identities_with_data_source=false: ${join(", ", local.unresolved_identities)}"
    }
  }
}

data "braintrustdata_user" "member_by_email" {
  for_each = local.identities_to_lookup

  email    = each.key
  org_name = var.org_name_for_user_lookup
}

locals {
  resolved_user_id_by_identity = merge(
    local.normalized_user_id_by_identity,
    {
      for identity, user in data.braintrustdata_user.member_by_email :
      identity => user.id
    }
  )

  team_member_user_ids = {
    for team_key, team in local.normalized_teams :
    team_key => sort(distinct(concat(
      team.member_user_ids,
      [
        for identity in team.member_identities :
        local.resolved_user_id_by_identity[identity]
      ]
    )))
  }
}

resource "braintrustdata_group" "team_groups" {
  for_each = local.normalized_teams

  name        = "${var.team_group_name_prefix}${each.value.name}"
  description = each.value.description

  member_users = local.team_member_user_ids[each.key]
  member_groups = [
    for member_team_key in each.value.member_team_keys :
    braintrustdata_group.team_groups[member_team_key].id
  ]

  depends_on = [terraform_data.validate_inputs]
}
