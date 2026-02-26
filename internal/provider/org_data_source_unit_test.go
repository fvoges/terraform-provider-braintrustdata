package provider

import (
	"errors"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSelectSingleOrganizationByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		orgName     string
		wantID      string
		orgs        []client.Organization
	}{
		"finds exact match": {
			orgs: []client.Organization{
				{ID: "org-a", Name: "other"},
				{ID: "org-b", Name: "target"},
			},
			orgName: "target",
			wantID:  "org-b",
		},
		"returns not found when no exact match": {
			orgs: []client.Organization{
				{ID: "org-a", Name: "other"},
			},
			orgName:     "target",
			wantErrType: errOrganizationNotFoundByName,
		},
		"returns multiple when ambiguous": {
			orgs: []client.Organization{
				{ID: "org-a", Name: "target"},
				{ID: "org-b", Name: "target"},
			},
			orgName:     "target",
			wantErrType: errMultipleOrganizationsFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			org, err := selectSingleOrganizationByName(tc.orgs, tc.orgName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if org == nil {
				t.Fatalf("expected organization, got nil")
			}
			if org.ID != tc.wantID {
				t.Fatalf("organization ID mismatch: got=%q want=%q", org.ID, tc.wantID)
			}
		})
	}
}

func TestPopulateOrgDataSourceModel(t *testing.T) {
	t.Parallel()

	model := OrgDataSourceModel{}
	org := &client.Organization{
		ID:                 "org-1",
		Name:               "Acme",
		APIURL:             stringPtr("https://api.acme.dev"),
		IsUniversalAPI:     boolPtr(true),
		IsDataplanePrivate: boolPtr(false),
		ProxyURL:           stringPtr("https://proxy.acme.dev"),
		RealtimeURL:        stringPtr("wss://realtime.acme.dev"),
		Created:            stringPtr("2026-02-26T00:00:00Z"),
		ImageRenderingMode: stringPtr("auto"),
	}

	populateOrgDataSourceModel(&model, org)

	if model.ID.ValueString() != "org-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.OrgID.ValueString() != "org-1" {
		t.Fatalf("org_id mismatch: got=%q", model.OrgID.ValueString())
	}
	if model.Name.ValueString() != "Acme" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.APIURL.ValueString() != "https://api.acme.dev" {
		t.Fatalf("api_url mismatch: got=%q", model.APIURL.ValueString())
	}
	if !model.IsUniversalAPI.ValueBool() {
		t.Fatalf("is_universal_api mismatch: got=%v", model.IsUniversalAPI.ValueBool())
	}
	if model.IsDataplanePrivate.ValueBool() {
		t.Fatalf("is_dataplane_private mismatch: got=%v", model.IsDataplanePrivate.ValueBool())
	}
	if model.ProxyURL.ValueString() != "https://proxy.acme.dev" {
		t.Fatalf("proxy_url mismatch: got=%q", model.ProxyURL.ValueString())
	}
	if model.RealtimeURL.ValueString() != "wss://realtime.acme.dev" {
		t.Fatalf("realtime_url mismatch: got=%q", model.RealtimeURL.ValueString())
	}
	if model.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.ImageRenderingMode.ValueString() != "auto" {
		t.Fatalf("image_rendering_mode mismatch: got=%q", model.ImageRenderingMode.ValueString())
	}
}

func TestPopulateOrgDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := OrgDataSourceModel{
		APIURL:             types.StringValue("stale"),
		IsUniversalAPI:     types.BoolValue(true),
		IsDataplanePrivate: types.BoolValue(true),
		ProxyURL:           types.StringValue("stale"),
		RealtimeURL:        types.StringValue("stale"),
		Created:            types.StringValue("stale"),
		ImageRenderingMode: types.StringValue("stale"),
	}
	org := &client.Organization{
		ID:   "org-2",
		Name: "Beta",
	}

	populateOrgDataSourceModel(&model, org)

	if !model.APIURL.IsNull() {
		t.Fatalf("expected api_url null")
	}
	if !model.IsUniversalAPI.IsNull() {
		t.Fatalf("expected is_universal_api null")
	}
	if !model.IsDataplanePrivate.IsNull() {
		t.Fatalf("expected is_dataplane_private null")
	}
	if !model.ProxyURL.IsNull() {
		t.Fatalf("expected proxy_url null")
	}
	if !model.RealtimeURL.IsNull() {
		t.Fatalf("expected realtime_url null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created null")
	}
	if !model.ImageRenderingMode.IsNull() {
		t.Fatalf("expected image_rendering_mode null")
	}
}
