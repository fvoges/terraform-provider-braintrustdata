# Data-Driven Access Control Module

This module manages Braintrust projects, role-like groups, project ACLs, and group membership using a single data model.

## Design Goals

- Data-first input model (maps/lists only).
- Stable keys for long-lived resources.
- Human-friendly membership input (`member_identities`) resolved to user IDs.
- Deterministic ACL expansion: one ACL per `(project, group, permission, restrict_object_type)`.

## Inputs

- `projects` (`map(object)`) creates `braintrustdata_project` resources.
- `groups` (`map(object)`) creates `braintrustdata_group` resources.
- `project_group_bindings` (`list(object)`) creates `braintrustdata_acl` resources on projects.
- `user_id_by_identity` (`map(string)`) resolves `groups[*].member_identities` to user IDs.

## Normalization Pattern

The module normalizes:

- identity keys (case-insensitive mode optional)
- whitespace in names/ids
- set/list ordering for deterministic plans
- ACL bindings into a flat canonical map (`acl_entries_by_key`)

Input validation fails fast when:

- `member_identities` are missing in `user_id_by_identity`
- `member_group_keys` references unknown groups
- `project_group_bindings` references unknown project/group keys

## Example

```hcl
module "access_control" {
  source = "../../modules/access-control-data-driven"

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
      description       = "Full control for fraud"
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
      project_key          = "recs"
      group_key            = "platform_admin"
      permissions          = ["read", "update"]
      restrict_object_type = "experiment"
    }
  ]

  user_id_by_identity = {
    "alice@example.com" = "user-111"
    "bob@example.com"   = "user-222"
    "eve@example.com"   = "user-333"
  }
}
```

## Plug-in Point for Upcoming User Data Sources

Today: feed `user_id_by_identity` from an external export.

When `braintrustdata_user` / `braintrustdata_users` data sources are available, resolve identity keys before calling this module and pass the resulting map into `user_id_by_identity`.

Example shape:

```hcl
# Pseudo-pattern for future provider data sources.
# data "braintrustdata_users" "directory" {
#   emails = local.all_member_identities
# }
#
# locals {
#   user_id_by_identity = {
#     for user in data.braintrustdata_users.directory.users :
#     lower(user.email) => user.id
#   }
# }
```

## Notes

- Keep `project` and `group` map keys stable; rename using `name` fields when needed.
- Avoid circular `member_group_keys` relationships.
- ACL entries are immutable in Braintrust; permission changes create replacement ACL resources.
