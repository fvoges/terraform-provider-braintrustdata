# braintrustdata_ai_secret Example

This folder contains runnable Terraform examples for `braintrustdata_ai_secret`.

`secret` is write-only: Braintrust does not return the raw value on read or import, so if you import an existing secret and later want to rotate it, you must re-supply `secret` in configuration. Whitespace-only secrets are rejected; non-empty secrets preserve any leading or trailing whitespace you intentionally provide.

Prerequisites:
- Terraform >= 1.4.0
- Environment variables: `BRAINTRUST_API_KEY` and `BRAINTRUST_ORG_ID` (recommended)

Files:
- `versions.tf`: Terraform and provider version contract
- `resource.tf`: example resource configuration

Run:
1. `cd examples/resources/braintrustdata_ai_secret`
2. `terraform init -backend=false`
3. `terraform validate`
4. `terraform plan`
