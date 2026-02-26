# Examples Index

This directory provides Terraform examples organized by lifecycle and complexity.

## Prerequisites

- Terraform >= 1.4.0
- Braintrust credentials in environment:
  - BRAINTRUST_API_KEY
  - BRAINTRUST_ORG_ID (recommended)

## Tiering

- Tier A (Quickstart): minimal, fast-start examples.
- Tier B (Composable Patterns): reusable building blocks with dependency wiring.
- Tier C (Workflow Architecture): multi-module workflows with ownership boundaries.

## Catalog

| Path | Tier | Purpose | Runtime Expectation |
| --- | --- | --- | --- |
| examples/provider | A | provider configuration and smoke test | Safe to run with credentials |
| examples/resources/* | A-B | resource creation patterns (minimal + practical) | May create resources in your org |
| examples/data-sources/* | A-B | lookup/filter patterns for existing objects | Requires existing objects for placeholder lookups |
| examples/workflows/access-control-data-driven | C | legacy all-in-one access-control workflow | Creates projects/groups/acls |
| examples/workflows/access-control-lifecycle | C | split-state lifecycle workflow (recommended for scale) | Two-state apply flow |
| examples/modules/* | B-C | reusable modules consumed by workflows | Module-level; use via workflows |

## Safe To Run vs Existing Objects Required

- Safe with only credentials:
  - examples/provider
- Usually requires existing objects or placeholder replacement:
  - examples/data-sources/*
- Creates/updates resources in Braintrust:
  - examples/resources/*
  - examples/workflows/*

## Version Contract

Runnable leaf examples under resources/ and data-sources/ include versions.tf with:

- required_version >= 1.4.0
- provider braintrustdata pinned to = 0.1.0

## Recommended Paths

- Start: examples/provider
- Resource basics: examples/resources/braintrustdata_project and examples/resources/braintrustdata_group
- Access control at scale: examples/workflows/access-control-lifecycle
- Legacy reference: examples/workflows/access-control-data-driven
