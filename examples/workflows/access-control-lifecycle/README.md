# Access Control Lifecycle Workflow (Recommended)

This workflow demonstrates split ownership across two Terraform states.

- teams-stack: owns teams and memberships.
- project-access-stack: owns projects, role groups, ACLs, and team-to-role-group bindings.

## Ownership Boundaries

| Concern | Owner |
| --- | --- |
| Team groups and member_users/member_groups | teams-stack |
| Projects | project-access-stack |
| Role-like project access groups | project-access-stack |
| ACL grants on projects | project-access-stack |
| Team->role-group membership wiring | project-access-stack (from teams output intent) |

## Apply Order

1. Apply teams-stack.
2. Apply project-access-stack.
3. Re-apply only the stack whose source model changed.

## Contract

- teams-stack exports role_group_member_group_ids_by_binding_key.
- project-access-stack consumes remote state output and validates binding delimiter compatibility.

## Related Modules

- examples/modules/team-membership-lifecycle
- examples/modules/project-access-lifecycle

## Legacy Reference

- examples/workflows/access-control-data-driven (all-in-one pattern)
