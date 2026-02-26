# Team Membership Lifecycle Module

This module owns team-group lifecycle and team memberships.

## Owns

- `braintrustdata_group` per team key
- team `member_users`
- team `member_groups` (nested teams)

## Does not own

- project role groups
- ACLs
- project resources

## Inputs

- `teams`
- `user_id_by_identity`
- `resolve_identities_with_data_source`
- `org_name_for_user_lookup`
- `identity_case_insensitive`
- `team_group_name_prefix`
- `binding_key_delimiter`

## Output contract

- `team_group_ids_by_key`
- `binding_key_delimiter`
- `role_group_member_group_ids_by_binding_key` (`project_key|role_key` => set(team_group_ids))

## Identity resolution order

1. `user_id_by_identity`
2. `data.braintrustdata_user` lookup by email (if enabled)
3. fail when unresolved and lookup disabled

## Delimiter contract

`binding_key_delimiter` must match the delimiter expected by `project-access-lifecycle`.
