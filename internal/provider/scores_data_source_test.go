package provider

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScoresDataSource_WithScoreNameAndProjectIDFilter(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	scoreID := os.Getenv("BRAINTRUST_SCORE_ID")
	if scoreID == "" {
		t.Skip("BRAINTRUST_SCORE_ID must be set for scores data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScoresDataSourceConfigWithScoreNameAndProjectIDFilter(scoreID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_scores.test", "scores.#", "1"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_scores.test", "scores.0.id", "data.braintrustdata_score.seed", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_scores.test", "scores.0.name", "data.braintrustdata_score.seed", "name"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_scores.test", "scores.0.project_id", "data.braintrustdata_score.seed", "project_id"),
					resource.TestCheckResourceAttr("data.braintrustdata_scores.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccScoresDataSourceConfigWithScoreNameAndProjectIDFilter(scoreID string) string {
	return `
data "braintrustdata_score" "seed" {
  id = "` + scoreID + `"
}

data "braintrustdata_scores" "test" {
  project_id = data.braintrustdata_score.seed.project_id
  score_name = data.braintrustdata_score.seed.name
  filter_ids = [data.braintrustdata_score.seed.id]
  depends_on = [data.braintrustdata_score.seed]
}
`
}

func TestAccScoresDataSource_InvalidPagination(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_scores" "test" {
  starting_after = "score-1"
  ending_before  = "score-2"
}
`,
				ExpectError: regexp.MustCompile(`(?i)cannot\s+specify\s+both.*starting_after.*ending_before`),
			},
		},
	})
}

func TestAccScoresDataSource_InvalidLimit(t *testing.T) {
	testAccScoreDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_scores" "test" {
  limit = 0
}
`,
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}
