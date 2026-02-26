package provider

import (
	"errors"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
)

func TestSelectSingleAPIKeyByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		apiKeyName  string
		wantID      string
		apiKeys     []client.APIKey
	}{
		"finds exact api key": {
			apiKeys: []client.APIKey{
				{ID: "api-key-a", Name: "other"},
				{ID: "api-key-b", Name: "target"},
			},
			apiKeyName: "target",
			wantID:     "api-key-b",
		},
		"returns not found when no exact match": {
			apiKeys: []client.APIKey{
				{ID: "api-key-a", Name: "other"},
			},
			apiKeyName:  "target",
			wantErrType: errAPIKeyNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			apiKeys: []client.APIKey{
				{ID: "api-key-a", Name: "target"},
				{ID: "api-key-b", Name: "target"},
			},
			apiKeyName:  "target",
			wantErrType: errMultipleAPIKeysFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			apiKey, err := selectSingleAPIKeyByName(tc.apiKeys, tc.apiKeyName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if apiKey == nil {
				t.Fatalf("expected api key, got nil")
			}
			if apiKey.ID != tc.wantID {
				t.Fatalf("expected api key ID %q, got %q", tc.wantID, apiKey.ID)
			}
		})
	}
}

func TestPopulateAPIKeyDataSourceModel(t *testing.T) {
	t.Parallel()

	model := APIKeyDataSourceModel{}
	apiKey := &client.APIKey{
		ID:          "api-key-1",
		Name:        "service-key",
		OrgID:       "org-1",
		PreviewName: "sk-1234",
		Created:     "2026-02-26T00:00:00Z",
		UserID:      "user-1",
		UserEmail:   "user@example.com",
	}

	populateAPIKeyDataSourceModel(&model, apiKey)

	if model.ID.ValueString() != "api-key-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "service-key" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.OrgID.ValueString() != "org-1" {
		t.Fatalf("org_id mismatch: got=%q", model.OrgID.ValueString())
	}
	if model.PreviewName.ValueString() != "sk-1234" {
		t.Fatalf("preview_name mismatch: got=%q", model.PreviewName.ValueString())
	}
	if model.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}
	if model.UserEmail.ValueString() != "user@example.com" {
		t.Fatalf("user_email mismatch: got=%q", model.UserEmail.ValueString())
	}
}

func TestPopulateAPIKeyDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := APIKeyDataSourceModel{}
	apiKey := &client.APIKey{
		ID:   "api-key-2",
		Name: "viewer-key",
	}

	populateAPIKeyDataSourceModel(&model, apiKey)

	if !model.OrgID.IsNull() {
		t.Fatalf("expected org_id to be null")
	}
	if !model.PreviewName.IsNull() {
		t.Fatalf("expected preview_name to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
	if !model.UserID.IsNull() {
		t.Fatalf("expected user_id to be null")
	}
	if !model.UserEmail.IsNull() {
		t.Fatalf("expected user_email to be null")
	}
}
