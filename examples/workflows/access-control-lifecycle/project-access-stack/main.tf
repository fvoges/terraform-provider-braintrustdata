provider "braintrustdata" {
  # Read credentials from environment by default:
  # BRAINTRUST_API_KEY
  # BRAINTRUST_ORG_ID
}

data "terraform_remote_state" "teams" {
  backend = "local"

  config = {
    path = "../teams-stack/terraform.tfstate"
  }
}

locals {
  binding_key_delimiter = "|"

  role_catalog = {
    admin = {
      description = "Project admin"
      permissions = ["read", "update", "delete", "read_acls", "update_acls"]
    }
    editor = {
      description = "Project editor"
      permissions = ["read", "update"]
    }
    viewer = {
      description = "Project viewer"
      permissions = ["read"]
    }
  }

  projects = {
    fraud = {
      name          = "fraud"
      description   = "Fraud detection"
      enabled_roles = ["admin", "editor", "viewer"]
    }

    recs = {
      name          = "recommendations"
      description   = "Recommendation engine"
      enabled_roles = ["admin", "editor", "viewer"]
      role_overrides = {
        viewer = {
          description          = "Recs read-only at experiment level"
          permissions          = ["read"]
          restrict_object_type = "experiment"
        }
      }
    }
  }
}

resource "terraform_data" "validate_remote_contract" {
  input = {
    local_binding_key_delimiter  = local.binding_key_delimiter
    remote_binding_key_delimiter = data.terraform_remote_state.teams.outputs.binding_key_delimiter
  }

  lifecycle {
    precondition {
      condition     = local.binding_key_delimiter == data.terraform_remote_state.teams.outputs.binding_key_delimiter
      error_message = "Binding delimiter mismatch between stacks. teams-stack and project-access-stack must use the same delimiter."
    }
  }
}

module "project_access" {
  source = "../../../modules/project-access-lifecycle"

  projects                                   = local.projects
  role_catalog                               = local.role_catalog
  role_group_member_group_ids_by_binding_key = data.terraform_remote_state.teams.outputs.role_group_member_group_ids_by_binding_key

  project_name_prefix    = "proj-"
  role_group_name_prefix = "access-"
  binding_key_delimiter  = local.binding_key_delimiter

  depends_on = [terraform_data.validate_remote_contract]
}

output "project_ids_by_key" {
  value = module.project_access.project_ids_by_key
}

output "role_group_ids_by_binding_key" {
  value = module.project_access.role_group_ids_by_binding_key
}
