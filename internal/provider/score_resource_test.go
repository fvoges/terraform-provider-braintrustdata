package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScoreResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScoreResourceConfig("test-score", "Initial score description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_score.test", "name", "test-score"),
					resource.TestCheckResourceAttr("braintrustdata_score.test", "score_type", "free-form"),
					resource.TestCheckResourceAttr("braintrustdata_score.test", "description", "Initial score description"),
					resource.TestCheckNoResourceAttr("braintrustdata_score.test", "categories"),
					resource.TestCheckResourceAttrSet("braintrustdata_score.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_score.test", "project_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_score.test", "created"),
				),
			},
			{
				Config: testAccScoreResourceConfig("test-score-updated", "Updated score description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_score.test", "name", "test-score-updated"),
					resource.TestCheckResourceAttr("braintrustdata_score.test", "description", "Updated score description"),
					resource.TestCheckNoResourceAttr("braintrustdata_score.test", "categories"),
				),
			},
			{
				ResourceName:      "braintrustdata_score.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccScoreResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-score-resource"
}

resource "braintrustdata_score" "test" {
  project_id   = braintrustdata_project.test.id
  name         = %[1]q
  score_type   = "free-form"
  description  = %[2]q
}
`, name, description)
}
