# Teams Stack

This stack owns team groups and team memberships.

## Inputs and Outputs

- Input model: teams map (member identities/user IDs/nested team keys + project role intents).
- Outputs:
  - team_group_ids_by_key
  - binding_key_delimiter
  - role_group_member_group_ids_by_binding_key

## Run

1. cd examples/workflows/access-control-lifecycle/teams-stack
2. terraform init
3. terraform apply

## Notes

- Outputs are consumed by project-access-stack via terraform_remote_state.
- This stack does not create projects or ACLs.
