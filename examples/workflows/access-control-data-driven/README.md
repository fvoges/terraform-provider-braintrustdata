# Access Control Workflow Example

This workflow shows how to drive Braintrust access control from data structures:

- projects
- role-like groups
- project permission bindings
- user/group memberships

## Run

```bash
cd examples/workflows/access-control-data-driven
terraform init
terraform plan
```

Credentials are read from environment variables:

- `BRAINTRUST_API_KEY`
- `BRAINTRUST_ORG_ID`

## Future User Lookup Integration

When user data sources are available, resolve `local.user_id_by_identity` from data sources and keep the module inputs unchanged.
