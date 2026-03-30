package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAISecretResource(t *testing.T) {
	name := fmt.Sprintf("tf-acc-ai-secret-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAISecretResourceConfig(name, "openai", "initial-secret-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("braintrustdata_ai_secret.test", "id"),
					resource.TestCheckResourceAttr("braintrustdata_ai_secret.test", "name", name),
					resource.TestCheckResourceAttr("braintrustdata_ai_secret.test", "type", "openai"),
					resource.TestCheckResourceAttr("braintrustdata_ai_secret.test", "secret", "initial-secret-value"),
					resource.TestCheckResourceAttrSet("braintrustdata_ai_secret.test", "org_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_ai_secret.test", "preview_secret"),
				),
			},
			{
				ResourceName:            "braintrustdata_ai_secret.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				Config: testAccAISecretResourceConfig(name, "anthropic", "rotated-secret-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_ai_secret.test", "name", name),
					resource.TestCheckResourceAttr("braintrustdata_ai_secret.test", "type", "anthropic"),
					resource.TestCheckResourceAttr("braintrustdata_ai_secret.test", "secret", "rotated-secret-value"),
				),
			},
		},
	})
}

func TestAccAISecretResource_MissingSecret(t *testing.T) {
	name := fmt.Sprintf("tf-acc-ai-secret-missing-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAISecretResourceConfigMissingSecret(name),
				ExpectError: regexp.MustCompile(`'secret' must be provided and non-empty`),
			},
		},
	})
}

func testAccAISecretResourceConfig(name, secretType, secret string) string {
	return fmt.Sprintf(`
resource "braintrustdata_ai_secret" "test" {
  name   = %[1]q
  type   = %[2]q
  secret = %[3]q
}
`, name, secretType, secret)
}

func testAccAISecretResourceConfigMissingSecret(name string) string {
	return fmt.Sprintf(`
resource "braintrustdata_ai_secret" "test" {
  name = %[1]q
  type = "openai"
}
`, name)
}
