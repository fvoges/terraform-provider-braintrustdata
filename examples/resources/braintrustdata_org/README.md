# braintrustdata_org Example

This folder contains runnable Terraform examples for braintrustdata_org.

Prerequisites:
- Terraform >= 1.4.0
- Environment variables: BRAINTRUST_API_KEY and BRAINTRUST_ORG_ID (recommended)

Files:
- versions.tf: Terraform and provider version contract
- resource.tf: example resource configuration
- import.sh: sample import command

Run:
1. cd examples/resources/braintrustdata_org
2. terraform init -backend=false
3. terraform validate
4. terraform plan

Notes:
- Replace placeholder IDs before apply.
- Manage only settings that your organization policy allows changing.
