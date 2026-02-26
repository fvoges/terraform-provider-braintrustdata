# Access Control Lifecycle Workflow

This workflow demonstrates split ownership across two Terraform states:

- `teams-stack`: owns teams and memberships
- `project-access-stack`: owns projects, role groups, ACLs, and team-to-role-group bindings

## Apply order

1. Apply `teams-stack`
2. Apply `project-access-stack`
