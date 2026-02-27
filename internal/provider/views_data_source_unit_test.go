package provider

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListViewsOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListViewsOptions
		model       ViewsDataSourceModel
	}{
		"builds all supported api-native filters": {
			model: ViewsDataSourceModel{
				ObjectID:      types.StringValue("project-1"),
				ObjectType:    types.StringValue("project"),
				FilterIDs:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("view-1"), types.StringValue("view-2")}),
				ViewName:      types.StringValue("default"),
				ViewType:      types.StringValue("projects"),
				StartingAfter: types.StringValue("view-10"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListViewsOptions{
				ObjectID:      "project-1",
				ObjectType:    client.ACLObjectTypeProject,
				IDs:           []string{"view-1", "view-2"},
				ViewName:      "default",
				ViewType:      client.ViewTypeProjects,
				StartingAfter: "view-10",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: ViewsDataSourceModel{
				ObjectID:      types.StringValue("project-1"),
				ObjectType:    types.StringValue("project"),
				StartingAfter: types.StringValue("view-1"),
				EndingBefore:  types.StringValue("view-2"),
			},
			wantErrLike: "Cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: ViewsDataSourceModel{
				ObjectID:   types.StringValue("project-1"),
				ObjectType: types.StringValue("project"),
				Limit:      types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
		"rejects empty object_id": {
			model: ViewsDataSourceModel{
				ObjectID:   types.StringValue(""),
				ObjectType: types.StringValue("project"),
			},
			wantErrLike: "'object_id' must be provided and non-empty",
		},
		"rejects whitespace object_type": {
			model: ViewsDataSourceModel{
				ObjectID:   types.StringValue("project-1"),
				ObjectType: types.StringValue(" "),
			},
			wantErrLike: "'object_type' must be provided and non-empty",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListViewsOptions(context.Background(), tc.model)
			if tc.wantErrLike != "" {
				if !diags.HasError() {
					t.Fatalf("expected diagnostic containing %q, got none", tc.wantErrLike)
				}
				found := false
				for _, diag := range diags {
					if strings.Contains(diag.Detail(), tc.wantErrLike) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected diagnostic containing %q, got %v", tc.wantErrLike, diags)
				}

				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if opts == nil {
				t.Fatalf("expected options, got nil")
			}

			if opts.ObjectID != tc.want.ObjectID ||
				opts.ObjectType != tc.want.ObjectType ||
				!reflect.DeepEqual(opts.IDs, tc.want.IDs) ||
				opts.ViewName != tc.want.ViewName ||
				opts.ViewType != tc.want.ViewType ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestProviderDataSourcesIncludeViewPair(t *testing.T) {
	t.Parallel()

	p, ok := New("test")().(*BraintrustProvider)
	if !ok {
		t.Fatalf("expected *BraintrustProvider")
	}

	dataSourceFactories := p.DataSources(context.Background())
	dataSourceNames := make(map[string]struct{}, len(dataSourceFactories))

	for _, factory := range dataSourceFactories {
		ds := factory()
		resp := &datasource.MetadataResponse{}
		ds.Metadata(context.Background(), datasource.MetadataRequest{
			ProviderTypeName: "braintrustdata",
		}, resp)

		dataSourceNames[resp.TypeName] = struct{}{}
	}

	if _, ok := dataSourceNames["braintrustdata_view"]; !ok {
		t.Fatalf("expected braintrustdata_view to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_views"]; !ok {
		t.Fatalf("expected braintrustdata_views to be registered")
	}
}
