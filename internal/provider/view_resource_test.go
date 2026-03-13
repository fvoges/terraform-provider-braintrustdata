package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccViewResource(t *testing.T) {
	resourceName := "braintrustdata_view.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccViewResourceConfig("codex-view-acc", "table", false, `{"search":{"filter":[],"match":[],"sort":[],"tag":[]}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "codex-view-acc"),
					resource.TestCheckResourceAttr(resourceName, "object_type", "project"),
					resource.TestCheckResourceAttr(resourceName, "view_type", "experiments"),
					resource.TestCheckResourceAttr(resourceName, "options", `{"freezeColumns":false,"viewType":"table"}`),
					resource.TestCheckResourceAttr(resourceName, "view_data", `{"search":{"filter":[],"match":[],"sort":[],"tag":[]}}`),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "object_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource %s not found in state", resourceName)
					}

					return fmt.Sprintf("%s,%s,%s",
						rs.Primary.ID,
						rs.Primary.Attributes["object_id"],
						rs.Primary.Attributes["object_type"],
					), nil
				},
				ImportStateVerify: true,
			},
			{
				Config: testAccViewResourceConfig("codex-view-acc-updated", "cards", true, `{"search":{"filter":[],"match":[{"key":"name","operator":"contains","value":"demo"}],"sort":[],"tag":[]}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "codex-view-acc-updated"),
					resource.TestCheckResourceAttr(resourceName, "options", `{"freezeColumns":true,"viewType":"cards"}`),
					resource.TestCheckResourceAttr(resourceName, "view_data", `{"search":{"filter":[],"match":[{"key":"name","operator":"contains","value":"demo"}],"sort":[],"tag":[]}}`),
				),
			},
		},
	})
}

func testAccViewResourceConfig(name, viewType string, freezeColumns bool, viewData string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-view"
}

resource "braintrustdata_view" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  view_type   = "experiments"
  name        = %[1]q
  options     = jsonencode({
    freezeColumns = %[2]t
    viewType      = %[3]q
  })
  view_data   = %[4]q
}
`, name, freezeColumns, viewType, viewData)
}
