provider "braintrustdata" {
  # Read credentials from environment by default:
  # BRAINTRUST_API_KEY
  # BRAINTRUST_ORG_ID
}

locals {
  binding_key_delimiter = "|"

  teams = {
    ml_team = {
      name              = "ml-team"
      description       = "ML team"
      member_identities = ["alice@example.com", "bob@example.com"]
      project_role_bindings = [
        { project_key = "fraud", role_key = "admin" },
        { project_key = "recs", role_key = "editor" },
      ]
    }

    analysts = {
      name              = "analysts"
      description       = "Analyst team"
      member_identities = ["eve@example.com"]
      project_role_bindings = [
        { project_key = "fraud", role_key = "viewer" },
      ]
    }
  }
}

module "teams" {
  source = "../../../modules/team-membership-lifecycle"

  teams = local.teams

  identity_case_insensitive           = true
  team_group_name_prefix              = "team-"
  binding_key_delimiter               = local.binding_key_delimiter
  resolve_identities_with_data_source = true
}

output "team_group_ids_by_key" {
  value = module.teams.team_group_ids_by_key
}

output "binding_key_delimiter" {
  value = module.teams.binding_key_delimiter
}

output "role_group_member_group_ids_by_binding_key" {
  value = module.teams.role_group_member_group_ids_by_binding_key
}
