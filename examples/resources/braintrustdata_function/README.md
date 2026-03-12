# braintrustdata_function Example

This folder contains runnable Terraform examples for braintrustdata_function.

Prerequisites:
- Terraform >= 1.4.0
- Environment variables: BRAINTRUST_API_KEY and BRAINTRUST_ORG_ID (recommended)

Files:
- versions.tf: Terraform and provider version contract
- resource.tf: example resource configuration
- import.sh (if present): sample import command

Run:
1. cd examples/resources/braintrustdata_function
2. terraform init -backend=false
3. terraform validate
4. terraform plan

Notes:
- Placeholder values are marked with: # replace with real ID or wire from data/resource
- If prerequisite objects do not exist, wire IDs from data sources/resources first.
- Treat `function_data`, `function_schema`, and `prompt_data` as sensitive inputs. Do not embed API keys or other secrets in those JSON payloads.
- Use `braintrustdata_environment_variable` to manage secret material that functions need at runtime.
