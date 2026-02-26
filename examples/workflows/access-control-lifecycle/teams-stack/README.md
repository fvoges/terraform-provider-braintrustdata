# Teams Stack

This stack owns team groups and team memberships.

## Run

```bash
cd examples/workflows/access-control-lifecycle/teams-stack
terraform init
terraform apply
```

Outputs from this stack are consumed by the project-access stack via remote state.
