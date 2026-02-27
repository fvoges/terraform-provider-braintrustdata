package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
)

func TestSelectSingleViewByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		viewName    string
		wantID      string
		views       []client.View
	}{
		"finds exact view": {
			views: []client.View{
				{ID: "view-a", Name: "other"},
				{ID: "view-b", Name: "target"},
			},
			viewName: "target",
			wantID:   "view-b",
		},
		"returns not found when no exact match": {
			views: []client.View{
				{ID: "view-a", Name: "other"},
			},
			viewName:    "target",
			wantErrType: errViewNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			views: []client.View{
				{ID: "view-a", Name: "target"},
				{ID: "view-b", Name: "target"},
			},
			viewName:    "target",
			wantErrType: errMultipleViewsFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			view, err := selectSingleViewByName(tc.views, tc.viewName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if view == nil {
				t.Fatalf("expected view, got nil")
			}
			if view.ID != tc.wantID {
				t.Fatalf("expected view ID %q, got %q", tc.wantID, view.ID)
			}
		})
	}
}

func TestPopulateViewDataSourceModel(t *testing.T) {
	t.Parallel()

	model := ViewDataSourceModel{}
	view := &client.View{
		ID:         "view-1",
		Name:       "default",
		ObjectID:   "project-1",
		ObjectType: client.ACLObjectTypeProject,
		ViewType:   client.ViewTypeProjects,
		Created:    "2026-02-27T00:00:00Z",
		DeletedAt:  "2026-02-28T00:00:00Z",
		UserID:     "user-1",
	}

	populateViewDataSourceModel(context.Background(), &model, view)

	if model.ID.ValueString() != "view-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "default" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.ObjectID.ValueString() != "project-1" {
		t.Fatalf("object_id mismatch: got=%q", model.ObjectID.ValueString())
	}
	if model.ObjectType.ValueString() != "project" {
		t.Fatalf("object_type mismatch: got=%q", model.ObjectType.ValueString())
	}
	if model.ViewType.ValueString() != "projects" {
		t.Fatalf("view_type mismatch: got=%q", model.ViewType.ValueString())
	}
	if model.Created.ValueString() != "2026-02-27T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.DeletedAt.ValueString() != "2026-02-28T00:00:00Z" {
		t.Fatalf("deleted_at mismatch: got=%q", model.DeletedAt.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}
}

func TestPopulateViewDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := ViewDataSourceModel{}
	view := &client.View{
		ID:         "view-2",
		Name:       "staging",
		ObjectID:   "project-2",
		ObjectType: client.ACLObjectTypeProject,
		ViewType:   client.ViewTypeProjects,
	}

	populateViewDataSourceModel(context.Background(), &model, view)

	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
	if !model.DeletedAt.IsNull() {
		t.Fatalf("expected deleted_at to be null")
	}
	if !model.UserID.IsNull() {
		t.Fatalf("expected user_id to be null")
	}
}
