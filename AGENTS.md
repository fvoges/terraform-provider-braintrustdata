# Repository Guidelines

## Project Structure & Module Organization
- `main.go`: provider entrypoint binary (`terraform-provider-braintrustdata`).
- `internal/client/`: Braintrust API client layer (HTTP models, CRUD methods, error handling).
- `internal/provider/`: Terraform Plugin Framework resources, data sources, and provider config.
- `docs/`: generated provider/resource/data source docs used by Terraform Registry publishing.
- `examples/`: runnable Terraform examples for provider config, resources, and data sources.
- Root configs: `GNUmakefile`, `.golangci.yml`, `.pre-commit-config.yaml`, `go.mod`.

## Build, Test, and Development Commands
- `make build`: compile provider binary in repo root.
- `make test`: run unit tests for `internal/client` and `internal/provider` with coverage.
- `make testacc`: run acceptance tests (`TF_ACC=1`) against real API credentials.
- `make fmt`: format Go code with `gofmt -w -s`.
- `make lint`: run `golangci-lint` (including `gosec`, `revive`, `staticcheck`).
- `make generate`: regenerate docs with `tfplugindocs`.
- `make pre-commit-run`: run all pre-commit hooks locally.

Example:
```bash
make fmt && make lint && make test
```

## Coding Style & Naming Conventions
- Language: Go (targeting Go 1.21+).
- Formatting: always run `make fmt`; do not manually align whitespace.
- Lint baseline: pass `make lint` with current `.golangci.yml` rules.
- Naming:
  - Go files: lowercase with underscores only when needed.
  - Tests: `*_test.go`, test funcs `TestXxx`.
  - Terraform docs/examples: match provider type names (e.g., `braintrustdata_dataset`).

## Testing Guidelines
- Framework: Go `testing` package (`go test`).
- Location: tests live beside implementation in `internal/client` and `internal/provider`.
- Prefer table-driven tests for validation, API error mapping, and schema behavior.
- Run `make test` before every PR; run `make testacc` when changing provider-resource/API integration behavior.

## Commit & Pull Request Guidelines
- Follow Conventional Commit style seen in history: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore(ci):`.
- Keep commits focused and descriptive; include scope when useful (example: `feat(experiments): ...`).
- PRs should include:
  - clear summary and motivation,
  - linked issue (`#123`) when applicable,
  - updated tests and docs (`make generate`) for behavior/schema changes.

## Security & Configuration Tips
- Never commit secrets; use `BRAINTRUST_API_KEY` and `BRAINTRUST_ORG_ID` environment variables.
- Pre-commit includes `gitleaks`; fix leaks before pushing.
