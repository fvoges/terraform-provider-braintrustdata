package provider

import (
	"reflect"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveOrgID(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		planOrgID    types.String
		clientOrgID  string
		wantID       string
		wantErrIsNil bool
	}{
		"uses plan org_id when set": {
			planOrgID:    types.StringValue("org-plan"),
			clientOrgID:  "org-client",
			wantID:       "org-plan",
			wantErrIsNil: true,
		},
		"falls back to provider org_id": {
			planOrgID:    types.StringNull(),
			clientOrgID:  "org-client",
			wantID:       "org-client",
			wantErrIsNil: true,
		},
		"errors when neither value is present": {
			planOrgID:    types.StringNull(),
			clientOrgID:  "",
			wantID:       "",
			wantErrIsNil: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveOrgID(tc.planOrgID, tc.clientOrgID)
			if (err == nil) != tc.wantErrIsNil {
				t.Fatalf("error mismatch: got err=%v", err)
			}
			if got != tc.wantID {
				t.Fatalf("id mismatch: got=%q want=%q", got, tc.wantID)
			}
		})
	}
}

func TestBuildOrganizationPatchRequestFromPlan(t *testing.T) {
	t.Parallel()

	plan := OrgResourceModel{
		Name:               types.StringValue("Acme"),
		APIURL:             types.StringValue("https://api.acme.dev"),
		IsUniversalAPI:     types.BoolValue(false),
		IsDataplanePrivate: types.BoolValue(true),
		ProxyURL:           types.StringValue("https://proxy.acme.dev"),
		RealtimeURL:        types.StringValue("wss://realtime.acme.dev"),
		ImageRenderingMode: types.StringValue("click_to_load"),
	}

	got, hasChanges := buildOrganizationPatchRequestFromPlan(plan)
	if !hasChanges {
		t.Fatalf("expected hasChanges=true")
	}

	want := &client.PatchOrganizationRequest{
		Name:               stringPtr("Acme"),
		APIURL:             stringPtr("https://api.acme.dev"),
		IsUniversalAPI:     boolPtr(false),
		IsDataplanePrivate: boolPtr(true),
		ProxyURL:           stringPtr("https://proxy.acme.dev"),
		RealtimeURL:        stringPtr("wss://realtime.acme.dev"),
		ImageRenderingMode: stringPtr("click_to_load"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("request mismatch:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestBuildOrganizationPatchRequestFromPlan_NoChanges(t *testing.T) {
	t.Parallel()

	plan := OrgResourceModel{
		Name:               types.StringUnknown(),
		APIURL:             types.StringNull(),
		IsUniversalAPI:     types.BoolUnknown(),
		IsDataplanePrivate: types.BoolNull(),
		ProxyURL:           types.StringUnknown(),
		RealtimeURL:        types.StringNull(),
		ImageRenderingMode: types.StringUnknown(),
	}

	got, hasChanges := buildOrganizationPatchRequestFromPlan(plan)
	if hasChanges {
		t.Fatalf("expected hasChanges=false")
	}
	if got != nil {
		t.Fatalf("expected nil request when no changes, got %#v", got)
	}
}

func TestBuildOrganizationPatchRequestFromDiff(t *testing.T) {
	t.Parallel()

	state := OrgResourceModel{
		Name:               types.StringValue("Acme"),
		APIURL:             types.StringValue("https://api.acme.dev"),
		IsUniversalAPI:     types.BoolValue(true),
		IsDataplanePrivate: types.BoolValue(false),
		ImageRenderingMode: types.StringValue("auto"),
	}

	plan := OrgResourceModel{
		Name:               types.StringValue("Acme Updated"),
		APIURL:             types.StringValue("https://api.acme.dev"),
		IsUniversalAPI:     types.BoolValue(false),
		IsDataplanePrivate: types.BoolValue(false),
		ImageRenderingMode: types.StringValue("blocked"),
	}

	got, hasChanges := buildOrganizationPatchRequestFromDiff(plan, state)
	if !hasChanges {
		t.Fatalf("expected hasChanges=true")
	}

	want := &client.PatchOrganizationRequest{
		Name:               stringPtr("Acme Updated"),
		IsUniversalAPI:     boolPtr(false),
		ImageRenderingMode: stringPtr("blocked"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("request mismatch:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestPopulateOrgResourceModel(t *testing.T) {
	t.Parallel()

	org := &client.Organization{
		ID:                 "org-123",
		Name:               "Acme",
		APIURL:             stringPtr("https://api.acme.dev"),
		IsUniversalAPI:     boolPtr(true),
		IsDataplanePrivate: boolPtr(false),
		ProxyURL:           stringPtr("https://proxy.acme.dev"),
		RealtimeURL:        stringPtr("wss://realtime.acme.dev"),
		Created:            stringPtr("2026-02-26T00:00:00Z"),
		ImageRenderingMode: stringPtr("auto"),
	}

	model := OrgResourceModel{}
	populateOrgResourceModel(&model, org, "org-123")

	if model.ID.ValueString() != "org-123" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.OrgID.ValueString() != "org-123" {
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
	if model.ImageRenderingMode.ValueString() != "auto" {
		t.Fatalf("image_rendering_mode mismatch: got=%q", model.ImageRenderingMode.ValueString())
	}
}

func TestPopulateOrgResourceModel_Nullables(t *testing.T) {
	t.Parallel()

	org := &client.Organization{ID: "org-123", Name: "Acme"}
	model := OrgResourceModel{}
	populateOrgResourceModel(&model, org, "org-123")

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

func stringPtr(v string) *string {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}
