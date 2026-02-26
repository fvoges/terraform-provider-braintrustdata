# Access Control Workflow Example (Legacy All-in-One)

This workflow drives projects, groups, ACLs, and membership from one consolidated model.

Status:
- Kept for backward reference.
- For scaled lifecycle ownership, prefer examples/workflows/access-control-lifecycle.

## What It Owns

- Projects
- Role-like groups
- Project ACL bindings
- Group memberships

## Run

1. cd examples/workflows/access-control-data-driven
2. terraform init
3. terraform plan

Credentials are read from environment:

- BRAINTRUST_API_KEY
- BRAINTRUST_ORG_ID

## Related

- Recommended split-state workflow: examples/workflows/access-control-lifecycle
- Legacy module used by this workflow: examples/modules/access-control-data-driven
