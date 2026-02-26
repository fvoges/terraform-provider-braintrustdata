default: build

# Build the provider
build:
	go build -o terraform-provider-braintrustdata

# Run unit tests
test:
	go test ./internal/client/... ./internal/provider/... -v -cover -timeout=120s

# Run acceptance tests
testacc:
	TF_ACC=1 go test ./internal/provider/... -v -count=1 -run '^TestAcc' -timeout=120m

# Install provider locally for testing
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/darwin_arm64
	cp terraform-provider-braintrustdata ~/.terraform.d/plugins/registry.terraform.io/braintrustdata/braintrustdata/0.1.0/darwin_arm64/

# Format Go code
fmt:
	gofmt -w -s .

# Run linter
lint:
	golangci-lint run
	$(MAKE) examples-lint

# Run examples static checks
examples-lint:
	bash scripts/check-examples.sh

# Generate documentation
generate:
	@if [ ! -f "$(shell which tfplugindocs)" ]; then \
		echo "Installing tfplugindocs..."; \
		go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest; \
	fi
	tfplugindocs generate

# Install pre-commit hooks
pre-commit-install:
	@if [ ! -f "$(shell which pre-commit)" ]; then \
		echo "pre-commit not found. Install with: pip install pre-commit"; \
		exit 1; \
	fi
	pre-commit install

# Run pre-commit hooks manually
pre-commit-run:
	pre-commit run --all-files

# Clean build artifacts
clean:
	rm -f terraform-provider-braintrustdata
	rm -rf dist/
	rm -f coverage.out

# Display help
help:
	@echo "Available targets:"
	@echo "  build              - Build the provider binary"
	@echo "  test               - Run unit tests"
	@echo "  testacc            - Run acceptance tests (requires TF_ACC=1)"
	@echo "  install            - Install provider locally for testing"
	@echo "  fmt                - Format Go code"
	@echo "  lint               - Run golangci-lint"
	@echo "  examples-lint      - Run static checks for examples/ integrity"
	@echo "  generate           - Generate documentation with tfplugindocs"
	@echo "  pre-commit-install - Install pre-commit hooks"
	@echo "  pre-commit-run     - Run pre-commit hooks manually"
	@echo "  clean              - Remove build artifacts"
	@echo "  help               - Display this help message"

.PHONY: build test testacc install fmt lint examples-lint generate pre-commit-install pre-commit-run clean help
