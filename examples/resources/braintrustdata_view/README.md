# braintrustdata_view Example

This folder contains runnable Terraform examples for braintrustdata_view.

Prerequisites:
- Terraform >= 1.4.0
- Environment variables: BRAINTRUST_API_KEY and BRAINTRUST_ORG_ID (recommended)

Files:
- versions.tf: Terraform and provider version contract
- resource.tf: example resource configuration
- import.sh: sample import command

Run:
1. cd examples/resources/braintrustdata_view
2. terraform init -backend=false
3. terraform validate
4. terraform plan

Notes:
- View import requires a composite ID: `<view_id>,<object_id>,<object_type>`
- This example creates a project first so the view has a valid owning object
