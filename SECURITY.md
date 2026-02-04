# Security Policy

## Reporting a Vulnerability

We take the security of the Terraform Provider for Braintrust seriously. If you believe you have found a security vulnerability, please report it to us responsibly.

### Reporting Process

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please email security reports to:

**security@braintrust.dev**

Include the following information:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)
- Your contact information

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
- **Assessment**: We will assess the vulnerability and determine its severity
- **Communication**: We will keep you informed of our progress
- **Fix**: We will develop and test a fix
- **Release**: We will release a patched version
- **Disclosure**: We will coordinate public disclosure with you

### Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Security Features

### Built-in Security

The provider implements multiple security controls:

1. **HTTPS-Only Communication**
   - All API requests use TLS 1.2+
   - http:// URLs are rejected at the client level
   - Certificate validation is enforced

2. **Sensitive Data Protection**
   - API keys marked as `Sensitive: true` in Terraform schema
   - Keys are redacted from logs and error messages
   - Automatic sanitization of sensitive data in diagnostics

3. **Secret Scanning**
   - Pre-commit hooks with gitleaks prevent accidental commits
   - CI/CD pipelines scan for exposed credentials
   - No secrets stored in code or version control

4. **Static Analysis**
   - gosec security scanner runs on every commit
   - CodeQL advanced vulnerability detection
   - golangci-lint with security-focused rules

5. **Dependency Security**
   - Dependabot monitors for vulnerable dependencies
   - Trivy scans for known CVEs
   - Automated dependency updates

6. **Supply Chain Security**
   - SBOM (Software Bill of Materials) generated for each release
   - GPG-signed releases
   - SHA256 checksums for binary verification

### Best Practices for Users

#### API Key Management

- **Never commit API keys to version control**
- Use environment variables: `BRAINTRUST_API_KEY`
- Store keys in a secret management system (HashiCorp Vault, AWS Secrets Manager, etc.)
- Rotate API keys regularly
- Use separate keys for different environments (dev/staging/prod)

#### Terraform State Security

Terraform state files contain sensitive data:

```terraform
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "braintrust/terraform.tfstate"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}
```

- **Always use remote state with encryption**
- Never commit `.tfstate` files
- Restrict access to state storage
- Enable versioning for state recovery

#### Access Control

- Use least-privilege API keys
- Create separate service accounts for Terraform
- Audit API key usage regularly
- Revoke unused keys immediately

#### Testing Security

Acceptance tests require API credentials:

- Use a **dedicated non-production organization**
- Create a **test-only service account** with minimal permissions
- Never use production API keys in CI/CD
- Rotate test credentials regularly

## Known Limitations

### API Key Lifecycle

- API keys are only visible during creation
- The provider cannot retrieve existing key values
- Key rotation requires destroying and recreating the resource
- Document key values securely immediately after creation

### Soft Deletes

Braintrust uses soft deletes for some resources:

- Deleted resources have `deleted_at` timestamp
- Resources may not be immediately removed from API
- Name conflicts may occur with soft-deleted resources
- Consider using unique naming strategies

### Rate Limiting

- Provider implements client-side rate limiting
- Retries use exponential backoff with jitter
- Maximum 3 retry attempts for rate-limited requests
- Concurrent operations may hit rate limits

## Security Testing

### Running Security Scans

```bash
# Install security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scanner
gosec ./...

# Run with SARIF output
gosec -fmt sarif -out results.sarif ./...
```

### Pre-commit Security Hooks

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
make pre-commit-install

# Run manually
pre-commit run --all-files
```

## Incident Response

If a security incident occurs:

1. **Immediate Actions**
   - Rotate all potentially compromised API keys
   - Review access logs for unauthorized activity
   - Isolate affected resources

2. **Assessment**
   - Determine scope of compromise
   - Identify affected resources
   - Document timeline of events

3. **Remediation**
   - Apply security patches
   - Update provider to latest version
   - Review and update security practices

4. **Communication**
   - Notify affected users
   - Provide remediation guidance
   - Publish incident report

## Security Roadmap

Future security enhancements:

- [ ] Support for HashiCorp Vault dynamic secrets
- [ ] OIDC/OAuth2 authentication support
- [ ] Enhanced audit logging
- [ ] Resource-level encryption
- [ ] Security compliance certifications

## Additional Resources

- [Braintrust Security](https://www.braintrust.dev/security)
- [Terraform Security Best Practices](https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)

## Contact

For security-related questions or concerns:

- Email: security@braintrust.dev
- Security advisories: [GitHub Security Advisories](https://github.com/braintrustdata/terraform-provider-braintrustdata/security/advisories)

---

Last updated: 2026-02-03
