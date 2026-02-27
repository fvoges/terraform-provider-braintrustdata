package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
)

func TestSelectSingleTagByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		tagName     string
		wantID      string
		tags        []client.Tag
	}{
		"finds exact tag": {
			tags: []client.Tag{
				{ID: "tag-a", Name: "other"},
				{ID: "tag-b", Name: "target"},
			},
			tagName: "target",
			wantID:  "tag-b",
		},
		"returns not found when no exact match": {
			tags: []client.Tag{
				{ID: "tag-a", Name: "other"},
			},
			tagName:     "target",
			wantErrType: errTagNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			tags: []client.Tag{
				{ID: "tag-a", Name: "target"},
				{ID: "tag-b", Name: "target"},
			},
			tagName:     "target",
			wantErrType: errMultipleTagsFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tag, err := selectSingleTagByName(tc.tags, tc.tagName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tag == nil {
				t.Fatalf("expected tag, got nil")
			}
			if tag.ID != tc.wantID {
				t.Fatalf("expected tag ID %q, got %q", tc.wantID, tag.ID)
			}
		})
	}
}

func TestPopulateTagDataSourceModel(t *testing.T) {
	t.Parallel()

	model := TagDataSourceModel{}
	tag := &client.Tag{
		ID:          "tag-1",
		Name:        "production",
		ProjectID:   "proj-1",
		UserID:      "user-1",
		Color:       "#0066CC",
		Description: "Production tag",
		Created:     "2026-02-26T00:00:00Z",
	}

	populateTagDataSourceModel(context.Background(), &model, tag)

	if model.ID.ValueString() != "tag-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "production" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.ProjectID.ValueString() != "proj-1" {
		t.Fatalf("project_id mismatch: got=%q", model.ProjectID.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}
	if model.Color.ValueString() != "#0066CC" {
		t.Fatalf("color mismatch: got=%q", model.Color.ValueString())
	}
	if model.Description.ValueString() != "Production tag" {
		t.Fatalf("description mismatch: got=%q", model.Description.ValueString())
	}
	if model.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
}

func TestPopulateTagDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := TagDataSourceModel{}
	tag := &client.Tag{
		ID:        "tag-2",
		Name:      "staging",
		ProjectID: "proj-2",
	}

	populateTagDataSourceModel(context.Background(), &model, tag)

	if !model.UserID.IsNull() {
		t.Fatalf("expected user_id to be null")
	}
	if !model.Color.IsNull() {
		t.Fatalf("expected color to be null")
	}
	if !model.Description.IsNull() {
		t.Fatalf("expected description to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
}
