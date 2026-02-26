# braintrustdata_project Example

This folder contains runnable Terraform examples for braintrustdata_project.

Prerequisites:
- Terraform >= 1.4.0
- Environment variables: BRAINTRUST_API_KEY and BRAINTRUST_ORG_ID (recommended)

Files:
- versions.tf: Terraform and provider version contract
- resource.tf: example resource configuration
- import.sh (if present): sample import command

Run:
1. cd examples/resources/braintrustdata_project
2. terraform init -backend=false
3. terraform validate
4. terraform plan

Notes:
- Placeholder values are marked with: # replace with real ID or wire from data/resource
- If prerequisite objects do not exist, wire IDs from data sources/resources first.
