package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScoreDataSource_ByID(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	scoreID := os.Getenv("BRAINTRUST_SCORE_ID")
	if scoreID == "" {
		t.Skip("BRAINTRUST_SCORE_ID must be set for score data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScoreDataSourceConfigByID(scoreID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_score.test", "id", scoreID),
					resource.TestCheckResourceAttrSet("data.braintrustdata_score.test", "name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_score.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_score.test", "score_type"),
				),
			},
		},
	})
}

func testAccScoreDataSourceConfigByID(scoreID string) string {
	return fmt.Sprintf(`
data "braintrustdata_score" "test" {
  id = %[1]q
}
`, scoreID)
}

func TestAccScoreDataSource_ByName(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	scoreID := os.Getenv("BRAINTRUST_SCORE_ID")
	if scoreID == "" {
		t.Skip("BRAINTRUST_SCORE_ID must be set for score data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScoreDataSourceConfigByName(scoreID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_score.by_name", "id", "data.braintrustdata_score.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_score.by_name", "name", "data.braintrustdata_score.by_id", "name"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_score.by_name", "project_id", "data.braintrustdata_score.by_id", "project_id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_score.by_name", "score_type", "data.braintrustdata_score.by_id", "score_type"),
				),
			},
		},
	})
}

func testAccScoreDataSourceConfigByName(scoreID string) string {
	return fmt.Sprintf(`
data "braintrustdata_score" "by_id" {
  id = %[1]q
}

data "braintrustdata_score" "by_name" {
  name       = data.braintrustdata_score.by_id.name
  project_id = data.braintrustdata_score.by_id.project_id
  score_type = data.braintrustdata_score.by_id.score_type
}
`, scoreID)
}

func TestAccScoreDataSource_MissingLookupAttributes(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data "braintrustdata_score" "test" {}`,
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func TestAccScoreDataSource_ConflictingLookupAttributes(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_score" "test" {
  id         = "score-1"
  name       = "quality"
  project_id = "proj-1"
}
`,
				ExpectError: regexp.MustCompile(`Cannot combine 'id' with searchable attributes`),
			},
		},
	})
}

func TestAccScoreDataSource_NotFound(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	missingName := "missing-score-ds-name-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "braintrustdata_score" "test" {
  name       = %q
  project_id = %q
}
`, missingName, "00000000-0000-0000-0000-000000000000"),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No score found with name: %s", missingName)),
			},
		},
	})
}

func testAccScoreDataSourceRequiresAPIKey(t *testing.T) {
	t.Helper()

	if os.Getenv("BRAINTRUST_API_KEY") == "" {
		t.Skip("BRAINTRUST_API_KEY must be set for acceptance testing")
	}
}
