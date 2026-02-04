# Dependency Version Tracking

This document tracks all tool and dependency versions used in this project. **Keep this updated!**

Last Updated: 2026-02-04

## Core Runtime

| Component | Current Version | Latest Stable | Status | Notes |
|-----------|----------------|---------------|--------|-------|
| Go | 1.25.6 | 1.25.6 | ✅ Current | Released Jan 2026 |
| Terraform (workflows) | 1.14.x | 1.14.4 | ✅ Current | Updated 2026-02-04 |

## Go Module Dependencies

Updated via: `go get -u ./... && go mod tidy`

| Package | Purpose | Auto-Updated |
|---------|---------|--------------|
| `github.com/hashicorp/terraform-plugin-framework` | Provider framework | ✅ Dependabot |
| `github.com/hashicorp/terraform-plugin-go` | Plugin SDK | ✅ Dependabot |
| `github.com/hashicorp/terraform-plugin-testing` | Testing framework | ✅ Dependabot |
| `golang.org/x/time` | Rate limiting | ✅ Dependabot |

## GitHub Actions

Updated via: Dependabot (`.github/dependabot.yml`)

| Action | Current | Latest | Status | Notes |
|--------|---------|--------|--------|-------|
| `actions/checkout` | v6 | v6 | ✅ Current | Updated 2026-02-04 |
| `actions/setup-go` | v6 | v6 | ✅ Current | Updated 2026-02-04, uses Node 24 |
| `golangci/golangci-lint-action` | v9 | v9 | ✅ Current | Updated 2026-02-04, pinned to v2.8.0 |
| `github/codeql-action/*` | v4 | v4 | ✅ Current | Updated 2026-02-04 |
| `hashicorp/setup-terraform` | v3 | v3.1.2 | ✅ Current | |
| `actions/dependency-review-action` | v4 | v4 | ✅ Current | |
| `codecov/codecov-action` | v4 | v4 | ✅ Current | |
| `crazy-max/ghaction-import-gpg` | v6 | v6 | ✅ Current | |
| `goreleaser/goreleaser-action` | v6 | v6 | ✅ Current | Updated 2026-02-04 |
| `softprops/action-gh-release` | v2 | v2 | ✅ Current | Updated 2026-02-04 |
| `securego/gosec` | v2.21.4 | v2.21.4 | ✅ Current | Pinned version 2026-02-04 |
| `aquasecurity/trivy-action` | 0.32.0 | 0.32.0 | ✅ Current | Pinned version 2026-02-04 |

## Development Tools

| Tool | Version | How to Check | How to Update |
|------|---------|--------------|---------------|
| golangci-lint | v2.8.0 | `golangci-lint version` | `brew install golangci-lint` or download from releases |
| tfplugindocs | latest | `tfplugindocs version` | `go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest` |
| pre-commit | latest | `pre-commit --version` | `pip install --upgrade pre-commit` |

**Note**: golangci-lint v2 requires proper configuration. See `.golangci.yml` with `version: "2"` field.

## Checking for Updates

### Quick Check
```bash
# Go version
go version

# Go modules
go list -m -u all

# GitHub Actions (manual check)
# Visit: https://github.com/marketplace

# Terraform
# Visit: https://releases.hashicorp.com/terraform/
```

### Full Audit
```bash
# Run comprehensive dependency audit
# (This command will be created as part of automation)
./scripts/audit-dependencies.sh
```

## Update Schedule

- **Weekly**: Automated Dependabot PRs for Go modules and GitHub Actions
- **Monthly**: Manual review of this file and comprehensive audit
- **Before releases**: Full dependency audit and updates
- **As needed**: Security updates applied immediately

## Version Pinning Policy

### When to Pin
- Security-critical tools (gosec, trivy) - pin to specific versions
- Production dependencies - use semantic versioning ranges
- Development tools - use @latest for convenience

### When NOT to Pin
- Avoid pinning to @master or branch names in CI
- Don't pin to outdated versions without documented reason

## Resources

- Go Releases: https://go.dev/doc/devel/release
- Terraform Releases: https://releases.hashicorp.com/terraform/
- GitHub Actions Marketplace: https://github.com/marketplace
- Go Module Updates: https://endoflife.date/go
- Terraform EOL: https://endoflife.date/terraform

## Notes

- This file should be reviewed and updated at the start of every significant development session
- When updating versions, run full test suite to ensure compatibility
- Document any version pins that are intentionally behind latest (e.g., for compatibility)
