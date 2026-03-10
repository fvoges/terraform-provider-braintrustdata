package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFunctionDataSource_ByID(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	functionID, ok := testAccFunctionLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_FUNCTION_ID must be set for function data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionDataSourceConfigByID(functionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_function.test", "id", functionID),
					resource.TestCheckResourceAttrSet("data.braintrustdata_function.test", "name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_function.test", "function_type"),
				),
			},
		},
	})
}

func testAccFunctionDataSourceConfigByID(functionID string) string {
	return fmt.Sprintf(`
data "braintrustdata_function" "test" {
  id = %q
}
`, functionID)
}

func TestAccFunctionDataSource_ByNameAndProject(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	functionID, ok := testAccFunctionLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_FUNCTION_ID must be set for function data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionDataSourceConfigByNameAndProject(functionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_function.by_name", "id", "data.braintrustdata_function.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_function.by_name", "project_id", "data.braintrustdata_function.by_id", "project_id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_function.by_name", "name", "data.braintrustdata_function.by_id", "name"),
				),
			},
		},
	})
}

func testAccFunctionDataSourceConfigByNameAndProject(functionID string) string {
	return fmt.Sprintf(`
data "braintrustdata_function" "by_id" {
  id = %q
}

data "braintrustdata_function" "by_name" {
  project_id = data.braintrustdata_function.by_id.project_id
  name       = data.braintrustdata_function.by_id.name
}
`, functionID)
}

func TestAccFunctionDataSource_BySlugAndProject(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	functionID, ok := testAccFunctionLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_FUNCTION_ID must be set for function data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionDataSourceConfigBySlugAndProject(functionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_function.by_slug", "id", "data.braintrustdata_function.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_function.by_slug", "project_id", "data.braintrustdata_function.by_id", "project_id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_function.by_slug", "slug", "data.braintrustdata_function.by_id", "slug"),
				),
			},
		},
	})
}

func testAccFunctionDataSourceConfigBySlugAndProject(functionID string) string {
	return fmt.Sprintf(`
data "braintrustdata_function" "by_id" {
  id = %q
}

data "braintrustdata_function" "by_slug" {
  project_id = data.braintrustdata_function.by_id.project_id
  slug       = data.braintrustdata_function.by_id.slug
}
`, functionID)
}

func TestAccFunctionDataSource_NotFound(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_function" "test" {
  id = "ffffffff-ffff-4fff-8fff-ffffffffffff"
}
`,
				ExpectError: regexp.MustCompile(`No function found with ID: ffffffff-ffff-4fff-8fff-ffffffffffff`),
			},
		},
	})
}

func TestAccFunctionDataSource_MissingProjectIDForName(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_function" "test" {
  name = "tool-a"
}
`,
				ExpectError: regexp.MustCompile(`'project_id' must be provided when using 'name' or 'slug'`),
			},
		},
	})
}

func TestAccFunctionDataSource_MissingProjectIDForSlug(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_function" "test" {
  slug = "tool-a"
}
`,
				ExpectError: regexp.MustCompile(`'project_id' must be provided when using 'name' or 'slug'`),
			},
		},
	})
}

func TestAccFunctionDataSource_ConflictingLookupAttributes(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_function" "test" {
  id         = "function-1"
  project_id = "project-1"
  name       = "tool-a"
}
`,
				ExpectError: regexp.MustCompile(`Cannot combine 'id' with searchable attributes`),
			},
		},
	})
}

func TestAccFunctionDataSource_ConflictingNameAndSlug(t *testing.T) {
	testAccFunctionDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_function" "test" {
  project_id = "project-1"
  name       = "tool-a"
  slug       = "tool-a"
}
`,
				ExpectError: regexp.MustCompile(`Cannot specify both 'name' and 'slug'`),
			},
		},
	})
}

func testAccFunctionLookupContext() (string, bool) {
	functionID := os.Getenv("BRAINTRUST_FUNCTION_ID")
	if functionID == "" {
		return "", false
	}

	return functionID, true
}

func testAccFunctionDataSourceRequiresAPIKey(t *testing.T) {
	t.Helper()

	if os.Getenv("BRAINTRUST_API_KEY") == "" {
		t.Skip("BRAINTRUST_API_KEY must be set for acceptance testing")
	}
}
