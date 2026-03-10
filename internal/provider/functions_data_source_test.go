package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFunctionsDataSource_WithLimit(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	functionID, ok := testAccFunctionLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_FUNCTION_ID must be set for functions data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionsDataSourceConfigWithLimit(functionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_functions.test", "functions.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_functions.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccFunctionsDataSourceConfigWithLimit(functionID string) string {
	return fmt.Sprintf(`
data "braintrustdata_function" "seed" {
  id = %q
}

data "braintrustdata_functions" "test" {
  limit      = 1
  depends_on = [data.braintrustdata_function.seed]
}
`, functionID)
}

func TestAccFunctionsDataSource_WithNameFilter(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	functionID, ok := testAccFunctionLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_FUNCTION_ID must be set for functions data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionsDataSourceConfigWithNameFilter(functionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.braintrustdata_functions.test", "functions.#"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_functions.test", "ids.#"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_functions.test", "functions.0.name", "data.braintrustdata_function.seed", "name"),
				),
			},
		},
	})
}

func testAccFunctionsDataSourceConfigWithNameFilter(functionID string) string {
	return fmt.Sprintf(`
data "braintrustdata_function" "seed" {
  id = %q
}

data "braintrustdata_functions" "test" {
  project_id = data.braintrustdata_function.seed.project_id
  name       = data.braintrustdata_function.seed.name
  limit      = 1
}
`, functionID)
}

func TestAccFunctionsDataSource_WithSlugFilter(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	functionID, ok := testAccFunctionLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_FUNCTION_ID must be set for functions data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionsDataSourceConfigWithSlugFilter(functionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_functions.test", "functions.#", "1"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_functions.test", "functions.0.id", "data.braintrustdata_function.seed", "id"),
				),
			},
		},
	})
}

func testAccFunctionsDataSourceConfigWithSlugFilter(functionID string) string {
	return fmt.Sprintf(`
data "braintrustdata_function" "seed" {
  id = %q
}

data "braintrustdata_functions" "test" {
  project_id = data.braintrustdata_function.seed.project_id
  slug       = data.braintrustdata_function.seed.slug
  limit      = 1
}
`, functionID)
}

func TestAccFunctionsDataSource_ZeroLimitAllowed(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_functions" "test" {
  limit = 0
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_functions.test", "limit", "0"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_functions.test", "ids.#"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_functions.test", "functions.#"),
				),
			},
		},
	})
}

func TestAccFunctionsDataSource_InvalidPagination(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_functions" "test" {
  starting_after = "function-1"
  ending_before  = "function-2"
}
`,
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func TestAccFunctionsDataSource_InvalidLimit(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_functions" "test" {
  limit = -1
}
`,
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 0`),
			},
		},
	})
}
