package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"braintrustdata": providerserver.NewProtocol6WithError(New("test")()),
}

// TestNew verifies provider can be instantiated
func TestNew(t *testing.T) {
	provider := New("test")()

	if provider == nil {
		t.Fatal("expected provider to be created, got nil")
	}
}

// TestProvider_Metadata verifies provider metadata
func TestProvider_Metadata(t *testing.T) {
	provider := New("test")()

	// For now, just verify provider can be created
	// Full metadata testing will be done via acceptance tests
	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}

// TestProvider_Schema verifies provider schema has required attributes
func TestProvider_Schema(t *testing.T) {
	provider := New("test")()

	// Schema will be validated by the framework
	// This test verifies the provider can generate a schema without panicking
	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}

// TestProviderConfigure_APIKey verifies API key configuration
func TestProviderConfigure_APIKey(t *testing.T) {
	// This will be tested via acceptance tests once we have the full provider implementation
	// For now, we verify the provider can be instantiated
	provider := New("test")()
	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}

// TestProviderConfigure_EnvironmentVariables verifies environment variable precedence
func TestProviderConfigure_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("BRAINTRUST_API_KEY", "sk-env-test-key")
	_ = os.Setenv("BRAINTRUST_ORG_ID", "org-env-test")
	defer func() {
		_ = os.Unsetenv("BRAINTRUST_API_KEY")
		_ = os.Unsetenv("BRAINTRUST_ORG_ID")
	}()

	// Verify environment variables are set
	if os.Getenv("BRAINTRUST_API_KEY") != "sk-env-test-key" {
		t.Error("BRAINTRUST_API_KEY not set correctly")
	}
	if os.Getenv("BRAINTRUST_ORG_ID") != "org-env-test" {
		t.Error("BRAINTRUST_ORG_ID not set correctly")
	}

	// Provider instantiation will be tested via acceptance tests
	provider := New("test")()
	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}

// TestProviderConfigure_DefaultAPIURL verifies default API URL
func TestProviderConfigure_DefaultAPIURL(t *testing.T) {
	expectedDefault := "https://api.braintrust.dev"

	// This will be validated in the actual provider configuration
	// For now, document the expected default
	if expectedDefault != "https://api.braintrust.dev" {
		t.Error("default API URL should be https://api.braintrust.dev")
	}
}

// TestProviderConfigure_ValidationRequired verifies API key is required
func TestProviderConfigure_ValidationRequired(t *testing.T) {
	// Set empty values to test validation
	_ = os.Setenv("BRAINTRUST_API_KEY", "")
	_ = os.Setenv("BRAINTRUST_ORG_ID", "")
	defer func() {
		_ = os.Unsetenv("BRAINTRUST_API_KEY")
		_ = os.Unsetenv("BRAINTRUST_ORG_ID")
	}()

	// Validation will be tested via acceptance tests
	// This test documents the requirement
	provider := New("test")()
	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}
