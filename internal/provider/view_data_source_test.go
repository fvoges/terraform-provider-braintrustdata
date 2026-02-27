package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccViewDataSource_ByID(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	viewID, objectID, objectType, ok := testAccViewLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_VIEW_ID, BRAINTRUST_VIEW_OBJECT_ID, and BRAINTRUST_VIEW_OBJECT_TYPE must be set for view data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccViewDataSourceConfigByID(viewID, objectID, objectType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_view.test", "id", viewID),
					resource.TestCheckResourceAttr("data.braintrustdata_view.test", "object_id", objectID),
					resource.TestCheckResourceAttr("data.braintrustdata_view.test", "object_type", objectType),
					resource.TestCheckResourceAttrSet("data.braintrustdata_view.test", "name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_view.test", "view_type"),
				),
			},
		},
	})
}

func testAccViewDataSourceConfigByID(viewID, objectID, objectType string) string {
	return fmt.Sprintf(`
data "braintrustdata_view" "test" {
  id          = %[1]q
  object_id   = %[2]q
  object_type = %[3]q
}
`, viewID, objectID, objectType)
}

func TestAccViewDataSource_ByName(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	viewID, objectID, objectType, ok := testAccViewLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_VIEW_ID, BRAINTRUST_VIEW_OBJECT_ID, and BRAINTRUST_VIEW_OBJECT_TYPE must be set for view data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccViewDataSourceConfigByName(viewID, objectID, objectType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_view.by_name", "id", "data.braintrustdata_view.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_view.by_name", "name", "data.braintrustdata_view.by_id", "name"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_view.by_name", "object_id", "data.braintrustdata_view.by_id", "object_id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_view.by_name", "object_type", "data.braintrustdata_view.by_id", "object_type"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_view.by_name", "view_type", "data.braintrustdata_view.by_id", "view_type"),
				),
			},
		},
	})
}

func testAccViewDataSourceConfigByName(viewID, objectID, objectType string) string {
	return fmt.Sprintf(`
data "braintrustdata_view" "by_id" {
  id          = %[1]q
  object_id   = %[2]q
  object_type = %[3]q
}

data "braintrustdata_view" "by_name" {
  name        = data.braintrustdata_view.by_id.name
  object_id   = data.braintrustdata_view.by_id.object_id
  object_type = data.braintrustdata_view.by_id.object_type
  view_type   = data.braintrustdata_view.by_id.view_type
}
`, viewID, objectID, objectType)
}

func TestAccViewDataSource_MissingLookupAttributes(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_view" "test" {
  object_id   = "project-1"
  object_type = "project"
}
`,
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func TestAccViewDataSource_ConflictingLookupAttributes(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_view" "test" {
  id          = "view-1"
  name        = "default"
  object_id   = "project-1"
  object_type = "project"
}
`,
				ExpectError: regexp.MustCompile(`Cannot combine 'id' with searchable attributes`),
			},
		},
	})
}

func TestAccViewDataSource_NotFound(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	_, objectID, objectType, ok := testAccViewLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_VIEW_OBJECT_ID and BRAINTRUST_VIEW_OBJECT_TYPE must be set for view data source acceptance testing")
	}

	missingName := "missing-view-ds-name-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "braintrustdata_view" "test" {
  name        = %q
  object_id   = %q
  object_type = %q
}
`, missingName, objectID, objectType),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No view found with name: %s", missingName)),
			},
		},
	})
}

func testAccViewLookupContext() (string, string, string, bool) {
	viewID := os.Getenv("BRAINTRUST_VIEW_ID")
	objectID := os.Getenv("BRAINTRUST_VIEW_OBJECT_ID")
	objectType := os.Getenv("BRAINTRUST_VIEW_OBJECT_TYPE")

	if viewID == "" || objectID == "" || objectType == "" {
		return "", "", "", false
	}

	return viewID, objectID, objectType, true
}

func testAccViewDataSourceRequiresAPIKey(t *testing.T) {
	t.Helper()

	if os.Getenv("BRAINTRUST_API_KEY") == "" {
		t.Skip("BRAINTRUST_API_KEY must be set for acceptance testing")
	}
}
