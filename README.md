# Terraform Provider for Braintrust

[![Tests](https://github.com/braintrustdata/terraform-provider-braintrustdata/actions/workflows/test.yml/badge.svg)](https://github.com/braintrustdata/terraform-provider-braintrustdata/actions/workflows/test.yml)
[![CodeQL](https://github.com/braintrustdata/terraform-provider-braintrustdata/actions/workflows/codeql.yml/badge.svg)](https://github.com/braintrustdata/terraform-provider-braintrustdata/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/braintrustdata/terraform-provider-braintrustdata)](https://goreportcard.com/report/github.com/braintrustdata/terraform-provider-braintrustdata)

Official Terraform provider for [Braintrust](https://braintrust.dev), enabling infrastructure-as-code management of projects, experiments, datasets, prompts, functions, and access control.

## Features

- **Project Management**: Create and manage Braintrust projects
- **Experiments & Datasets**: Define experiments and datasets for ML evaluation
- **Prompt Templates**: Version-controlled prompt management with automatic versioning
- **Custom Functions**: Deploy scorer functions, tools, and tasks
- **Access Control**: Fine-grained permissions with groups, ACLs, and API keys
- **Full CRUD Support**: Complete lifecycle management for all resources
- **Import Existing Resources**: Bring existing infrastructure under Terraform management
- **Secure by Default**: HTTPS-only, TLS 1.2+, sensitive data sanitization

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for development)
- Braintrust API key (obtain from [braintrust.dev](https://www.braintrust.dev))

## Installation

### Terraform Registry (Recommended)

```terraform
terraform {
  required_providers {
    braintrustdata = {
      source  = "braintrustdata/braintrustdata"
      version = "~> 0.1"
    }
  }
}

provider "braintrustdata" {
  api_key         = "sk-***"  # Or use BRAINTRUST_API_KEY env var
  organization_id = "org-***" # Or use BRAINTRUST_ORG_ID env var
}
```

### Local Development

```bash
make install
```

## Quick Start

```terraform
# Create a project
resource "braintrustdata_project" "example" {
  name        = "My ML Project"
  description = "Evaluation project for my ML model"
}

# Create a dataset
resource "braintrustdata_dataset" "example" {
  project_id  = braintrustdata_project.example.id
  name        = "Test Dataset"
  description = "Test cases for evaluation"
}

# Create an experiment
resource "braintrustdata_experiment" "example" {
  project_id = braintrustdata_project.example.id
  name       = "Baseline Evaluation"
  dataset_id = braintrustdata_dataset.example.id
  public     = false
}
```

## Authentication

The provider requires a Braintrust API key. You can provide it in three ways:

1. **Provider configuration**:
   ```terraform
   provider "braintrustdata" {
     api_key = "sk-***"
   }
   ```

2. **Environment variable** (recommended):
   ```bash
   export BRAINTRUST_API_KEY="sk-***"
   export BRAINTRUST_ORG_ID="org-***"
   ```

3. **Terraform variables**:
   ```terraform
   variable "braintrust_api_key" {
     type      = string
     sensitive = true
   }

   provider "braintrustdata" {
     api_key = var.braintrust_api_key
   }
   ```

⚠️ **Never commit API keys to version control.** Use environment variables or a secret management system.

## Documentation

Comprehensive documentation is available in the [docs](./docs) directory:

- [Provider Configuration](./docs/index.md)
- [Resources](./docs/resources/)
- [Data Sources](./docs/data-sources/)
- [Examples](./examples/)

## Development

### Prerequisites

```bash
# Install Go dependencies
go mod download

# Install development tools
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install pre-commit hooks
pip install pre-commit
make pre-commit-install
```

### Building

```bash
make build
```

### Testing

```bash
# Run unit tests
make test

# Run acceptance tests (requires API credentials)
export BRAINTRUST_API_KEY="sk-***"
export BRAINTRUST_ORG_ID="org-***"
make testacc
```

### Running Acceptance Tests on PRs (Maintainers Only)

Acceptance tests normally only run on main branch pushes. To run them on a PR:

1. Go to Actions tab in GitHub
2. Select "Tests" workflow
3. Click "Run workflow"
4. Select the PR branch
5. Check "Run acceptance tests" checkbox
6. Click "Run workflow"

Note: This requires maintainer permissions and uses repository secrets.

### Test-Driven Development

This provider is built using strict TDD methodology:

1. Write failing tests first (RED)
2. Implement minimal code to pass (GREEN)
3. Refactor while keeping tests green (REFACTOR)

All new features must follow this process. See [superpowers/test-driven-development](https://github.com/anthropics/claude-code) for details.

### API Reference

The Braintrust API is documented via OpenAPI specification at https://github.com/braintrustdata/braintrust-openapi. Consult this specification when implementing new resources or debugging API interactions.

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Generate documentation
make generate
```

### Pre-commit Hooks

The project uses pre-commit hooks to ensure code quality:

- **gofmt**: Code formatting
- **go vet**: Static analysis
- **gosec**: Security scanning
- **gitleaks**: Secret detection
- **go test**: Run tests
- **tfplugindocs**: Generate documentation

Hooks run automatically on `git commit`. To run manually:

```bash
make pre-commit-run
```

## Security

### Reporting Vulnerabilities

Please report security vulnerabilities to [security@braintrust.dev](mailto:security@braintrust.dev). See [SECURITY.md](./SECURITY.md) for details.

### Security Features

- **HTTPS-only**: All API communication uses TLS 1.2+
- **Sensitive data sanitization**: API keys redacted from logs and errors
- **Secret scanning**: gitleaks prevents accidental credential commits
- **Static analysis**: gosec and CodeQL scan for vulnerabilities
- **Dependency scanning**: Dependabot and Trivy monitor dependencies
- **SBOM generation**: Software Bill of Materials for supply chain transparency

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Follow TDD: Write tests first
4. Ensure all tests pass: `make test`
5. Run linters: `make lint`
6. Generate docs: `make generate`
7. Submit a pull request

## Release Process

1. Update `CHANGELOG.md` with changes
2. Run pre-release validation:
   ```bash
   # Via GitHub Actions
   gh workflow run pre-release-check.yml -f version=v0.2.0
   ```
3. After checks pass, create and push tag:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```
4. GitHub Actions will automatically build and publish the release

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](./LICENSE) file for details.

## Resources

- [Braintrust Documentation](https://www.braintrust.dev/docs)
- [Braintrust API Reference](https://www.braintrust.dev/docs/api-reference)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Provider Registry](https://registry.terraform.io/providers/braintrustdata/braintrustdata/latest)

## Support

- [GitHub Issues](https://github.com/braintrustdata/terraform-provider-braintrustdata/issues)
- [Braintrust Community](https://www.braintrust.dev/community)
- [Documentation](https://www.braintrust.dev/docs)
