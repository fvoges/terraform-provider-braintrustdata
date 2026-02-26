# braintrustdata_project Example

This folder contains runnable Terraform examples for braintrustdata_project.

Prerequisites:
- Terraform >= 1.4.0
- Environment variables: BRAINTRUST_API_KEY and BRAINTRUST_ORG_ID (recommended)

Files:
- versions.tf: Terraform and provider version contract
- data-source.tf: example data-source lookups and outputs

Run:
1. cd examples/data-sources/braintrustdata_project
2. terraform init -backend=false
3. terraform validate
4. terraform plan

Notes:
- Placeholder values are marked with: # replace with real ID or wire from data/resource
- Data sources perform live API reads during planning.
