# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Project resource with full CRUD support
- Dataset resource with project references
- Experiment resource with dataset and project references
- Prompt resource with versioning
- Function resource for custom scorers and tools
- Project score resource for evaluation metrics
- Group resource for user collections
- ACL resource for fine-grained permissions
- API key resource for authentication management
- Data sources for all resources
- Import support for existing infrastructure
- Comprehensive acceptance tests

## [0.1.0] - TBD

### Added
- Initial provider implementation following HashiCorp Plugin Framework
- Provider configuration with API key authentication
- HTTPS-only enforcement with TLS 1.2+
- Environment variable support (BRAINTRUST_API_KEY, BRAINTRUST_ORG_ID, BRAINTRUST_API_URL)
- API client with Bearer token authentication
- Rate limiting and retry logic with exponential backoff
- Structured error handling with sensitive data sanitization
- Comprehensive test suite with TDD methodology
- Security scanning (gosec, CodeQL, Trivy, gitleaks)
- Pre-commit hooks for code quality and security
- CI/CD workflows (test, lint, docs, pre-release-check, release)
- Dependabot for dependency management
- Multi-platform builds (darwin/linux/windows, amd64/arm64)
- GPG signing of releases
- SBOM generation for supply chain transparency
- Documentation automation with tfplugindocs

### Security
- HTTPS-only API communication (http:// URLs rejected)
- TLS 1.2+ with certificate validation
- Sensitive attribute masking for API keys
- Error message sanitization to prevent data leaks
- Secret scanning with gitleaks
- Static analysis with gosec and CodeQL
- Dependency vulnerability scanning with Trivy and Dependabot
- Test environment isolation with dedicated non-production org

## Release Guidelines

### Before Each Release

1. Update this CHANGELOG.md with new version section
2. Run pre-release validation workflow
3. Verify all security scans pass
4. Ensure documentation is up to date
5. Validate multi-platform builds

### Version Numbering

- **Major** (X.0.0): Breaking changes to provider schema or behavior
- **Minor** (0.X.0): New resources, data sources, or features (backward compatible)
- **Patch** (0.0.X): Bug fixes and minor improvements (backward compatible)

[Unreleased]: https://github.com/braintrustdata/terraform-provider-braintrustdata/compare/v0.1.0...HEAD
