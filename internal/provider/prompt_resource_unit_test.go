package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestSlugify verifies the slug derivation helper.
func TestSlugify(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  string
	}{
		{input: "My Prompt", want: "my-prompt"},
		{input: "support agent", want: "support-agent"},
		{input: "Hello World!", want: "hello-world"},
		{input: "already-slug", want: "already-slug"},
		{input: "MixedCase123", want: "mixedcase123"},
		{input: "special@#chars", want: "specialchars"},
		{input: "  leading-trailing  ", want: "--leading-trailing--"},
		{input: "a b  c", want: "a-b--c"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := slugify(tc.input)
			if got != tc.want {
				t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestBuildCreatePromptRequest_SlugDerivedFromName verifies that when slug is
// not provided, it is derived from the name.
func TestBuildCreatePromptRequest_SlugDerivedFromName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	data := PromptResourceModel{
		ProjectID:   types.StringValue("project-123"),
		Name:        types.StringValue("My Support Agent"),
		Slug:        types.StringNull(),
		Description: types.StringNull(),
		Tags:        types.SetValueMust(types.StringType, []attr.Value{}),
		Metadata:    types.MapNull(types.StringType),
		PromptData:  types.StringNull(),
	}

	req, diags := buildCreatePromptRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.Slug != "my-support-agent" {
		t.Errorf("expected derived slug %q, got %q", "my-support-agent", req.Slug)
	}
}

// TestBuildCreatePromptRequest_ExplicitSlugWins verifies that when slug is
// explicitly provided, it is used as-is rather than derived from the name.
func TestBuildCreatePromptRequest_ExplicitSlugWins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	data := PromptResourceModel{
		ProjectID:   types.StringValue("project-123"),
		Name:        types.StringValue("My Support Agent"),
		Slug:        types.StringValue("custom-slug"),
		Description: types.StringNull(),
		Tags:        types.SetValueMust(types.StringType, []attr.Value{}),
		Metadata:    types.MapNull(types.StringType),
		PromptData:  types.StringNull(),
	}

	req, diags := buildCreatePromptRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.Slug != "custom-slug" {
		t.Errorf("expected explicit slug %q, got %q", "custom-slug", req.Slug)
	}
}

func TestBuildUpdatePromptRequest_SendsExplicitClearFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := PromptResourceModel{
		Name:         types.StringValue("support-agent-v2"),
		Slug:         types.StringValue("support-agent-v2"),
		Description:  types.StringNull(),
		FunctionType: types.StringNull(),
		Metadata:     types.MapNull(types.StringType),
		Tags:         types.SetNull(types.StringType),
		PromptData:   types.StringNull(),
	}

	req, diags := buildUpdatePromptRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal update payload: %v", err)
	}

	for _, key := range []string{"description", "function_type", "metadata", "tags", "prompt_data"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("expected %q to be present in update payload when config sets explicit clear", key)
		}
	}
}

func TestBuildUpdatePromptRequest_OmitsUnknownOptionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := PromptResourceModel{
		Name:         types.StringValue("support-agent-v2"),
		Slug:         types.StringUnknown(),
		Description:  types.StringUnknown(),
		FunctionType: types.StringUnknown(),
		Metadata:     types.MapUnknown(types.StringType),
		Tags:         types.SetUnknown(types.StringType),
		PromptData:   types.StringUnknown(),
	}

	req, diags := buildUpdatePromptRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal update payload: %v", err)
	}

	if _, ok := payload["name"]; !ok {
		t.Fatal("expected name to be included in update payload")
	}

	for _, key := range []string{"slug", "description", "function_type", "metadata", "tags", "prompt_data"} {
		if _, ok := payload[key]; ok {
			t.Errorf("expected %q to be omitted when value is unknown", key)
		}
	}
}

// TestSetPromptResourceModel_NullTagsStaysNull verifies that when tags is
// omitted from config (plan carries null), setPromptResourceModel preserves
// null after the API returns nil/empty tags. Previously it always wrote an
// empty set, causing plan=null vs state=empty-set inconsistency errors.
func TestSetPromptResourceModel_NullTagsStaysNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name string
		tags []string
	}{
		{name: "nil_tags", tags: nil},
		{name: "empty_tags", tags: []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prompt := &client.Prompt{
				ID:        "prompt-1",
				ProjectID: "project-1",
				Name:      "test",
				Tags:      tc.tags,
			}

			// Simulate plan/state where tags was omitted from config (null).
			var data PromptResourceModel
			data.Tags = types.SetNull(types.StringType)

			diags := setPromptResourceModel(ctx, &data, prompt)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if !data.Tags.IsNull() {
				t.Errorf("expected null tags (omitted from config), got non-null: %v", data.Tags)
			}
		})
	}
}

// TestSetPromptResourceModel_EmptyTagsBecomesEmptySet verifies that when the
// API returns nil or empty tags, setPromptResourceModel writes an empty set
// (not null) to state. This prevents a perpetual diff when the config
// declares `tags = []`.
func TestSetPromptResourceModel_EmptyTagsBecomesEmptySet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name string
		tags []string
	}{
		{name: "nil_tags", tags: nil},
		{name: "empty_tags", tags: []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prompt := &client.Prompt{
				ID:        "prompt-1",
				ProjectID: "project-1",
				Name:      "test",
				Tags:      tc.tags,
			}

			// Simulate plan/state where config has explicit `tags = []`.
			var data PromptResourceModel
			data.Tags = types.SetValueMust(types.StringType, []attr.Value{})

			diags := setPromptResourceModel(ctx, &data, prompt)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if data.Tags.IsNull() {
				t.Errorf("expected non-null tags set, got null; config tags=[] would produce perpetual diff")
			}
			if data.Tags.IsUnknown() {
				t.Errorf("expected known tags set, got unknown")
			}

			// Should be an empty set (zero elements).
			elems := data.Tags.Elements()
			if len(elems) != 0 {
				t.Errorf("expected 0 tag elements, got %d", len(elems))
			}
		})
	}
}

// TestSetPromptResourceModel_NonEmptyTagsRoundTrip verifies that non-empty tags
// are preserved correctly through setPromptResourceModel.
func TestSetPromptResourceModel_NonEmptyTagsRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	prompt := &client.Prompt{
		ID:        "prompt-1",
		ProjectID: "project-1",
		Name:      "test",
		Tags:      []string{"ml", "production"},
	}

	var data PromptResourceModel
	diags := setPromptResourceModel(ctx, &data, prompt)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if data.Tags.IsNull() || data.Tags.IsUnknown() {
		t.Fatalf("expected known non-null tags, got null=%v unknown=%v", data.Tags.IsNull(), data.Tags.IsUnknown())
	}

	elems := data.Tags.Elements()
	if len(elems) != 2 {
		t.Errorf("expected 2 tag elements, got %d", len(elems))
	}

	// Verify the expected values are present.
	found := map[string]bool{}
	for _, e := range elems {
		sv, ok := e.(types.String)
		if !ok {
			t.Fatalf("expected types.String element, got %T", e)
		}
		found[sv.ValueString()] = true
	}
	for _, want := range []string{"ml", "production"} {
		if !found[want] {
			t.Errorf("expected tag %q not found in set", want)
		}
	}
}

// TestSetPromptResourceModel_EmptyTagsMatchesConfigEmptySet verifies that
// an empty set produced by setPromptResourceModel is equal to the empty set
// that Terraform would produce from `tags = []` in config. This is the
// critical equality check that prevents perpetual diffs.
func TestSetPromptResourceModel_EmptyTagsMatchesConfigEmptySet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	prompt := &client.Prompt{
		ID:        "prompt-1",
		ProjectID: "project-1",
		Name:      "test",
		Tags:      nil,
	}

	// Simulate plan/state where config has explicit `tags = []`.
	var data PromptResourceModel
	data.Tags = types.SetValueMust(types.StringType, []attr.Value{})

	diags := setPromptResourceModel(ctx, &data, prompt)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// Simulate what Terraform would put in state from `tags = []`.
	configEmptySet := types.SetValueMust(types.StringType, []attr.Value{})

	if !data.Tags.Equal(configEmptySet) {
		t.Errorf("state tags %v is not equal to config empty set %v; this will produce a perpetual diff", data.Tags, configEmptySet)
	}
}

// TestSetPromptResourceModel_PromptDataMarshalError verifies that a marshal
// failure on prompt_data produces a Terraform diagnostic error and returns
// early rather than silently setting prompt_data to null.
//
// Note: json.Marshal on a standard interface{} value rarely fails in practice.
// We simulate the failure by passing a value that cannot be marshalled — in
// this case a channel type embedded inside the interface.  We achieve this by
// constructing a Prompt whose PromptData field is an unmarshalable Go type
// (a map containing a function value, which json.Marshal cannot encode).
func TestSetPromptResourceModel_PromptDataMarshalError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// json.Marshal fails for values containing channels, functions, or
	// complex numbers.  We use a map[string]interface{} containing a
	// function value to trigger the failure.
	unmarshalable := map[string]interface{}{
		"key": func() {}, // functions are not JSON-serialisable
	}

	prompt := &client.Prompt{
		ID:         "prompt-1",
		ProjectID:  "project-1",
		Name:       "test",
		PromptData: unmarshalable,
	}

	var data PromptResourceModel
	diags := setPromptResourceModel(ctx, &data, prompt)

	if !diags.HasError() {
		t.Fatal("expected a diagnostic error from marshal failure, but got none")
	}

	// Verify the error summary matches what we document.
	found := false
	for _, d := range diags {
		if d.Summary() == "Error Encoding prompt_data" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected diagnostic with summary %q, got: %v", "Error Encoding prompt_data", diags)
	}
}

// TestSetPromptResourceModel_PromptDataNilBecomesNull verifies that a nil
// PromptData in the API response maps to a null types.String in state.
func TestSetPromptResourceModel_PromptDataNilBecomesNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	prompt := &client.Prompt{
		ID:         "prompt-1",
		ProjectID:  "project-1",
		Name:       "test",
		PromptData: nil,
	}

	var data PromptResourceModel
	diags := setPromptResourceModel(ctx, &data, prompt)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !data.PromptData.IsNull() {
		t.Errorf("expected null prompt_data for nil API response, got %v", data.PromptData)
	}
}

// TestSetPromptResourceModel_PromptDataRoundTrip verifies that valid
// prompt_data is JSON-encoded and stored as a string value in state.
func TestSetPromptResourceModel_PromptDataRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	promptData := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.7,
	}

	prompt := &client.Prompt{
		ID:         "prompt-1",
		ProjectID:  "project-1",
		Name:       "test",
		PromptData: promptData,
	}

	var data PromptResourceModel
	diags := setPromptResourceModel(ctx, &data, prompt)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if data.PromptData.IsNull() || data.PromptData.IsUnknown() {
		t.Fatalf("expected non-null/non-unknown prompt_data, got null=%v unknown=%v",
			data.PromptData.IsNull(), data.PromptData.IsUnknown())
	}

	if data.PromptData.ValueString() == "" {
		t.Error("expected non-empty JSON string for prompt_data")
	}
}
