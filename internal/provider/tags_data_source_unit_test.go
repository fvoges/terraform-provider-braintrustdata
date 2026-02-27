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

func TestBuildListTagsOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListTagsOptions
		model       TagsDataSourceModel
	}{
		"builds all supported API-native filters": {
			model: TagsDataSourceModel{
				FilterIDs:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("tag-1"), types.StringValue("tag-2")}),
				OrgName:       types.StringValue("example-org"),
				ProjectID:     types.StringValue("proj-1"),
				ProjectName:   types.StringValue("example-project"),
				TagName:       types.StringValue("production"),
				StartingAfter: types.StringValue("tag-10"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListTagsOptions{
				IDs:           []string{"tag-1", "tag-2"},
				OrgName:       "example-org",
				ProjectID:     "proj-1",
				ProjectName:   "example-project",
				TagName:       "production",
				StartingAfter: "tag-10",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: TagsDataSourceModel{
				StartingAfter: types.StringValue("tag-1"),
				EndingBefore:  types.StringValue("tag-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: TagsDataSourceModel{
				Limit: types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListTagsOptions(context.Background(), tc.model)
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

			if !reflect.DeepEqual(opts.IDs, tc.want.IDs) ||
				opts.OrgName != tc.want.OrgName ||
				opts.ProjectID != tc.want.ProjectID ||
				opts.ProjectName != tc.want.ProjectName ||
				opts.TagName != tc.want.TagName ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestProviderDataSourcesIncludeTagPair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_tag"]; !ok {
		t.Fatalf("expected braintrustdata_tag to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_tags"]; !ok {
		t.Fatalf("expected braintrustdata_tags to be registered")
	}
}
