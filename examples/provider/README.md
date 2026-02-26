# Provider Example

This example validates provider configuration and connectivity with a lightweight data-source read.

## Prerequisites

- Terraform `>= 1.4.0`
- Environment variables:
  - `BRAINTRUST_API_KEY`
  - `BRAINTRUST_ORG_ID` (recommended)

## Files

- `versions.tf`: Terraform and provider version constraints.
- `provider.tf`: Provider configuration + smoke-test data source.

## Run

```bash
cd examples/provider
terraform init -backend=false
terraform validate
terraform plan
```

If configuration is valid, `plan` will evaluate `data.braintrustdata_users.smoke` and output `smoke_user_ids`.

## Notes

- This is a connectivity/configuration check, not an infrastructure provisioning example.
- For provisioning examples, see `examples/resources/*` and `examples/workflows/*`.
