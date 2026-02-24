provider "braintrustdata" {
  # Read credentials from environment by default:
  # BRAINTRUST_API_KEY
  # BRAINTRUST_ORG_ID
}

locals {
  access_model = {
    projects = {
      fraud = {
        name        = "fraud"
        description = "Fraud detection"
      }
      recs = {
        name        = "recommendations"
        description = "Recommendation engine"
      }
    }

    groups = {
      fraud_admin = {
        name              = "proj-fraud-admin"
        description       = "Can administer the fraud project"
        member_identities = ["alice@example.com", "bob@example.com"]
      }

      fraud_reader = {
        name              = "proj-fraud-reader"
        member_identities = ["eve@example.com"]
      }

      platform_admin = {
        name            = "platform-admin"
        member_user_ids = ["866a8a8a-fee9-4a5b-8278-12970de499c2"]
      }
    }

    project_group_bindings = [
      {
        project_key = "fraud"
        group_key   = "fraud_admin"
        permissions = ["read", "update", "delete", "update_acls"]
      },
      {
        project_key = "fraud"
        group_key   = "fraud_reader"
        permissions = ["read"]
      },
      {
        project_key = "recs"
        group_key   = "platform_admin"
        permissions = ["read", "update"]
      }
    ]
  }

  # Current path: supply this map from your upstream identity system export.
  user_id_by_identity = {
    "alice@example.com" = "user-111"
    "bob@example.com"   = "user-222"
    "eve@example.com"   = "user-333"
  }
}

module "access_control" {
  source = "../../modules/access-control-data-driven"

  projects               = local.access_model.projects
  groups                 = local.access_model.groups
  project_group_bindings = local.access_model.project_group_bindings
  user_id_by_identity    = local.user_id_by_identity

  project_name_prefix       = "prod-"
  group_name_prefix         = "prod-"
  identity_case_insensitive = true
}

output "project_ids_by_key" {
  value = module.access_control.project_ids_by_key
}

output "group_ids_by_key" {
  value = module.access_control.group_ids_by_key
}
