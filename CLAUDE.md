# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for Braintrust (braintrust.dev) built using the Terraform Plugin Framework. The provider enables infrastructure-as-code management of Braintrust resources including projects, experiments, datasets, prompts, functions, and access control.

**Current Status**: Phase 1 complete (foundation + authentication), Phase 2 complete (Group resource + data sources).

## 🚨 CRITICAL RULES

### Never Mention Claude in Files or Commits

**NEVER** include any reference to "Claude", "Claude Code", "Claude Sonnet", "Anthropic", or AI assistance in:
- Git commit messages (including Co-Authored-By)
- Code comments
- Documentation files
- Any file checked into the repository

User wants commits and code to appear as if written by a human developer without AI assistance.

### Coordinator Role - Never Code Directly

**CRITICAL: Your role is COORDINATION ONLY. NEVER write code directly.**

**You MUST delegate all coding work to specialized sub-agents:**
- ✅ **DO**: Identify issues, launch sub-agents (Task tool), review output, manage PRs
- ❌ **NEVER**: Use Edit, Write, or directly modify implementation files
- ❌ **NEVER**: Fix bugs, add features, or write code yourself

**Process for ANY code change:**
1. Launch general-purpose sub-agent using Task tool
2. Provide clear task description and context
3. Let the sub-agent write/edit code
4. Review their work
5. Coordinate next steps

**Examples of what requires sub-agents:**
- Fixing bugs or lint errors
- Adding features or tests
- Modifying any `.go`, `.tf`, or implementation files
- Creating new code files

**Exceptions (you can do directly):**
- Updating MEMORY.md or CLAUDE.md (documentation/process)
- Managing GitHub issues, PRs, and review conversations
- Running bash commands for git, gh CLI, status checks

**This is non-negotiable.** Even for "quick fixes" - always delegate to sub-agents.

### Dependency Currency

**ALWAYS ensure all tools and dependencies use current recommended versions.**

Before making ANY changes to the codebase, CI/CD, or tooling:

### 1. **Verify Current Versions**
Check that these are using the latest stable/recommended versions:
- Go version in `go.mod`
- Go module dependencies in `go.mod` and `go.sum`
- GitHub Actions in `.github/workflows/*.yml`
- Terraform version in workflows
- Tool versions: golangci-lint, tfplugindocs, pre-commit hooks

### 2. **When to Check**
- **Every session start**: Quick scan of critical dependencies
- **Before modifying workflows**: Full audit of GitHub Actions
- **Before adding dependencies**: Verify latest version of new deps
- **Monthly**: Comprehensive dependency review (can be automated)

### 3. **How to Check**
```bash
# Go version
go version
# Check against: https://go.dev/doc/devel/release

# Go modules
go list -m -u all
# Update: go get -u ./... && go mod tidy

# GitHub Actions versions
# Search: "actions/checkout latest version"
# Search: "golangci-lint-action latest version"
# Check: https://github.com/marketplace

# Terraform version
# Check: https://releases.hashicorp.com/terraform/
```

### 4. **Update Policy**
- ✅ **Always use latest stable** for runtime (Go, Terraform)
- ✅ **Always use latest** for CI/CD actions (GitHub Actions)
- ✅ **Pin specific versions** (not @master) for security tools
- ⚠️ **Test before updating** major versions of frameworks
- 📝 **Document** version choices in comments when pinning to older versions

### 5. **Automation**
- Dependabot configured for GitHub Actions and Go modules
- Renovate bot can be added for more comprehensive updates
- Monthly manual review of all dependencies

**Failure to maintain current versions leads to:**
- Security vulnerabilities
- CI/CD failures (like the golangci-lint Go version mismatch)
- Compatibility issues
- Technical debt accumulation

## Local Development Setup

### Required Tools

Install these tools before starting development:

```bash
# Go (required)
# Download from https://go.dev/dl/ or use version manager

# Terraform (required for testing examples)
brew install terraform

# golangci-lint (required for linting)
brew install golangci-lint

# tfplugindocs (required for documentation generation)
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

# pre-commit (optional but recommended)
pip install pre-commit
make pre-commit-install
```

### Verify Installation

```bash
go version          # Should be 1.25.6+
terraform version   # Should be 1.14.x+
golangci-lint version  # Should be v2.8.0+
tfplugindocs version   # Should show usage
```

## Essential Commands

### Development Workflow
```bash
# Build provider binary
make build

# Run unit tests (client + provider tests)
make test

# Run specific test
go test ./internal/provider/... -run TestGroupResource_Create -v

# Run acceptance tests (requires API credentials)
export BRAINTRUST_API_KEY="sk-***"
export BRAINTRUST_ORG_ID="org-***"
make testacc

# Run specific acceptance test
TF_ACC=1 go test ./internal/provider/... -run TestAccGroupResource -v -count=1

# Format code
make fmt

# Run linters
make lint

# Generate documentation
make generate

# Install provider locally for manual testing
make install
```

### Manual Testing
```bash
# Test provider installation
cd examples/provider
terraform init
terraform plan

# Test group resource
cd examples/resources/braintrustdata_group
terraform init
terraform apply
```

## API Specification

The Braintrust API is documented via OpenAPI specification:
https://github.com/braintrustdata/braintrust-openapi

This specification defines all API endpoints, request/response schemas, and field types used by the Terraform provider. When adding new resources, always consult the OpenAPI spec to ensure accurate field types and validation.

## Architecture

### Two-Layer Design

**Layer 1: API Client** (`internal/client/`)
- Raw HTTP client for Braintrust API
- Enforces HTTPS-only with TLS 1.2+ (panics on http://)
- Handles authentication via Bearer tokens
- Resource-specific methods (e.g., `CreateGroup`, `GetGroup`, `UpdateGroup`, `DeleteGroup`)
- Returns Go structs, not Terraform types

**Layer 2: Terraform Provider** (`internal/provider/`)
- Terraform Plugin Framework implementation
- Maps Terraform configurations to API calls
- Resource implementations (e.g., `GroupResource`)
- Converts between Terraform types (`types.String`, `types.List`) and Go types
- Handles Terraform state management and plan modifications

### Key Architectural Patterns

1. **Security-First Client Design**
   - `NewClient()` enforces HTTPS and panics on http:// URLs
   - TLS 1.2+ configured via `tls.Config{MinVersion: tls.VersionTLS12}`
   - Sensitive data (API keys) sanitized from error messages

2. **Error Handling**
   - Client returns typed errors: `NotFoundError`, `ValidationError`, `RateLimitError`, `APIError`
   - Provider converts client errors to Terraform diagnostics
   - Resources check for `NotFoundError` to handle deleted resources gracefully

3. **Resource Lifecycle Pattern**
   - Each resource implements: `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`, `ImportState`
   - `Configure` receives the API client from provider
   - CRUD methods convert between Terraform models and client structs
   - `Read` with missing resource returns no error (resource removed from state)

4. **Provider Configuration Precedence**
   - Priority: explicit config → environment variables → defaults
   - Example: `api_key` config > `BRAINTRUST_API_KEY` env var
   - Organization ID defaults to provider's `organization_id` if not specified on resource

### File Organization

```
internal/
├── client/              # API client layer
│   ├── client.go       # Core HTTP client, NewClient, Do()
│   ├── errors.go       # Typed error definitions
│   ├── groups.go       # Group-specific API methods
│   └── *_test.go       # Client unit tests
└── provider/            # Terraform provider layer
    ├── provider.go      # Provider implementation, Configure()
    ├── group_resource.go        # Group resource implementation
    └── *_resource_test.go       # Resource acceptance tests
```

### Test Organization

- **Client tests** (`internal/client/*_test.go`): Test API client in isolation with mocked HTTP responses
- **Provider tests** (`internal/provider/provider_test.go`): Test provider configuration and setup
- **Resource tests** (`internal/provider/*_resource_test.go`): Acceptance tests that require `TF_ACC=1` and real API credentials

## Development Guidelines

### Test-Driven Development (TDD)

This project follows strict TDD methodology:
1. **RED**: Write failing test first (unit test for client, acceptance test for resource)
2. **GREEN**: Implement minimal code to pass the test
3. **REFACTOR**: Improve code while keeping tests green

All new features must follow this process. See README.md section on TDD for details.

### Adding a New Resource

1. **Client Layer** (TDD in `internal/client/`)
   - Define Go structs for API request/response
   - Write tests for CRUD operations in `<resource>_test.go`
   - Implement methods: `Create<Resource>`, `Get<Resource>`, `Update<Resource>`, `Delete<Resource>`

2. **Provider Layer** (TDD in `internal/provider/`)
   - Define `<Resource>ResourceModel` struct with `tfsdk` tags
   - Write acceptance tests in `<resource>_resource_test.go`
   - Implement resource with all required methods
   - Register resource in `provider.go` `Resources()` method

3. **Documentation**
   - Add examples in `examples/resources/braintrustdata_<resource>/`
   - Run `make generate` to create docs from schema
   - Optionally add templates in `templates/` for custom documentation

### Adding Tests

Run tests with verbose output to see detailed logs:
```bash
# Unit tests with coverage
go test ./internal/client/... -v -cover

# Specific test with detailed output
go test ./internal/provider/... -run TestGroupResource_Create -v

# Acceptance tests (use -count=1 to disable test caching)
TF_ACC=1 go test ./internal/provider/... -v -count=1 -timeout=120m
```

### Common Patterns

**Converting Terraform List to Go Slice:**
```go
var memberIDs []string
data.MemberIDs.ElementsAs(ctx, &memberIDs, false)
```

**Converting Go Slice to Terraform List:**
```go
memberIDsList, diags := types.ListValueFrom(ctx, types.StringType, group.MemberIDs)
resp.Diagnostics.Append(diags...)
```

**Handling Resource Not Found:**
```go
if errors.As(err, &client.NotFoundError{}) {
    resp.State.RemoveResource(ctx)
    return
}
```

**Default to Provider Organization ID:**
```go
orgID := r.client.OrgID()
if !data.OrgID.IsNull() {
    orgID = data.OrgID.ValueString()
}
```

## Security Considerations

- Never commit API keys or sensitive data
- Use environment variables for credentials: `BRAINTRUST_API_KEY`, `BRAINTRUST_ORG_ID`
- Client enforces HTTPS-only (panics on http://)
- API keys are marked `Sensitive: true` in provider schema
- Pre-commit hooks include gitleaks for secret detection
- CodeQL and gosec scan for security vulnerabilities in CI

## CI/CD

Pre-commit hooks automatically run on commit:
- Code formatting (gofmt)
- Static analysis (go vet)
- Security scanning (gosec, gitleaks)
- Tests (go test)
- Documentation generation (tfplugindocs)

GitHub Actions workflows:
- `test.yml`: Run tests on push/PR
- `codeql.yml`: Security analysis
- `release.yml`: Automated releases on version tags
- `pre-release-check.yml`: Pre-release validation

## Authentication for Testing

To test locally or run acceptance tests:

```bash
export BRAINTRUST_API_KEY="sk-***"  # Get from braintrust.dev settings
export BRAINTRUST_ORG_ID="org-***"  # Your organization ID
```

The provider supports three configuration methods (in precedence order):
1. Provider configuration block
2. Environment variables (recommended)
3. Terraform variables

## Debugging Provider

Run provider in debug mode:
```bash
go build -o terraform-provider-braintrustdata
./terraform-provider-braintrustdata -debug
```

Then use the output `TF_REATTACH_PROVIDERS` environment variable with Terraform commands.
