# Project Access Lifecycle Module

This module owns project lifecycle and project-scoped access roles.

## Owns

- `braintrustdata_project` per project key
- `braintrustdata_group` per `(project_key, role_key)`
- `braintrustdata_acl` per `(project_key, role_key, permission, restrict_object_type)`
- role-group `member_groups` bindings from `role_group_member_group_ids_by_binding_key`

## Inputs

- `projects`
- `role_catalog`
- `role_group_member_group_ids_by_binding_key`
- `project_name_prefix`
- `role_group_name_prefix`
- `binding_key_delimiter`

## Output contract

- `project_ids_by_key`
- `role_group_ids_by_binding_key`
- `role_group_names_by_binding_key`

## Notes

- Binding key format is `project_key|role_key` by default.
- This module is the sole writer of project role groups and their `member_groups`.
