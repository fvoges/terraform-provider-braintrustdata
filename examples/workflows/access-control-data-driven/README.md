# Access Control Workflow Example

This workflow shows how to drive Braintrust access control from data structures:

- projects
- role-like groups
- project permission bindings
- user/group memberships

It resolves email identities to user IDs using the `braintrustdata_user` data source.

## Run

```bash
cd examples/workflows/access-control-data-driven
terraform init
terraform plan
```

Credentials are read from environment variables:

- `BRAINTRUST_API_KEY`
- `BRAINTRUST_ORG_ID`
