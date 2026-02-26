# Project Access Stack

This stack owns projects, role groups, ACLs, and role-group memberships.

## Inputs and Dependencies

- Reads role_group_member_group_ids_by_binding_key from ../teams-stack/terraform.tfstate.
- Validates binding delimiter contract before applying module wiring.

## Run

1. cd examples/workflows/access-control-lifecycle/project-access-stack
2. terraform init
3. terraform apply

## Apply Order

1. Apply teams-stack first.
2. Apply project-access-stack second.
3. Re-apply only the stack whose source model changed.

## Notes

- This stack is the sole writer of role-group member_groups bindings.
- Unknown binding keys from remote-state intent fail fast via module validation.
