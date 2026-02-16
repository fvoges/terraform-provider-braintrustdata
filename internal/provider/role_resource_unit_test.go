package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestComputeStringSliceDiff(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		current []string
		desired []string
		wantAdd []string
		wantDel []string
	}{
		{
			name:    "no_changes",
			current: []string{"read", "update"},
			desired: []string{"read", "update"},
			wantAdd: nil,
			wantDel: nil,
		},
		{
			name:    "adds_and_removes",
			current: []string{"read", "delete"},
			desired: []string{"read", "update"},
			wantAdd: []string{"update"},
			wantDel: []string{"delete"},
		},
		{
			name:    "adds_all_when_current_empty",
			current: nil,
			desired: []string{"read", "update"},
			wantAdd: []string{"read", "update"},
			wantDel: nil,
		},
		{
			name:    "removes_all_when_desired_empty",
			current: []string{"read", "update"},
			desired: nil,
			wantAdd: nil,
			wantDel: []string{"read", "update"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotAdd, gotDel := computeStringSliceDiff(tc.current, tc.desired)
			if !reflect.DeepEqual(gotAdd, tc.wantAdd) {
				t.Fatalf("computeStringSliceDiff() add mismatch: got=%v want=%v", gotAdd, tc.wantAdd)
			}
			if !reflect.DeepEqual(gotDel, tc.wantDel) {
				t.Fatalf("computeStringSliceDiff() remove mismatch: got=%v want=%v", gotDel, tc.wantDel)
			}
		})
	}
}

func TestListToStringSlice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name   string
		input  types.List
		want   []string
		hasErr bool
	}{
		{
			name:  "null_list",
			input: types.ListNull(types.StringType),
			want:  nil,
		},
		{
			name: "known_values",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringValue("update"),
			}),
			want: []string{"read", "update"},
		},
		{
			name: "null_and_unknown_elements_are_filtered",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue("delete"),
			}),
			want: []string{"read", "delete"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, diags := listToStringSlice(ctx, tc.input)
			if tc.hasErr != diags.HasError() {
				t.Fatalf("listToStringSlice() diagnostics mismatch: hasErr=%v diags=%v", tc.hasErr, diags)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("listToStringSlice() mismatch: got=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestListToStringSliceWithState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name      string
		input     types.List
		want      []string
		wantState listValueState
		hasErr    bool
	}{
		{
			name:      "null_list",
			input:     types.ListNull(types.StringType),
			want:      nil,
			wantState: listValueStateNull,
		},
		{
			name:      "unknown_list",
			input:     types.ListUnknown(types.StringType),
			want:      nil,
			wantState: listValueStateUnknown,
		},
		{
			name: "known_values",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringValue("update"),
			}),
			want:      []string{"read", "update"},
			wantState: listValueStateKnown,
		},
		{
			name: "known_values_with_null_and_unknown_elements_filtered",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue("delete"),
			}),
			want:      []string{"read", "delete"},
			wantState: listValueStateKnown,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotState, diags := listToStringSliceWithState(ctx, tc.input)
			if tc.hasErr != diags.HasError() {
				t.Fatalf("listToStringSliceWithState() diagnostics mismatch: hasErr=%v diags=%v", tc.hasErr, diags)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("listToStringSliceWithState() mismatch: got=%v want=%v", got, tc.want)
			}
			if gotState != tc.wantState {
				t.Fatalf("listToStringSliceWithState() state mismatch: got=%v want=%v", gotState, tc.wantState)
			}
		})
	}
}

func TestComputeStringSliceDiffForDesiredState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		current      []string
		desired      []string
		wantAdd      []string
		wantDel      []string
		desiredState listValueState
	}{
		{
			name:         "desired_unknown_skips_diff",
			current:      []string{"read", "update"},
			desired:      nil,
			desiredState: listValueStateUnknown,
			wantAdd:      nil,
			wantDel:      nil,
		},
		{
			name:         "desired_null_removes_existing",
			current:      []string{"read", "update"},
			desired:      nil,
			desiredState: listValueStateNull,
			wantAdd:      nil,
			wantDel:      []string{"read", "update"},
		},
		{
			name:         "desired_known_uses_regular_diff",
			current:      []string{"read"},
			desired:      []string{"read", "update"},
			desiredState: listValueStateKnown,
			wantAdd:      []string{"update"},
			wantDel:      nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotAdd, gotDel := computeStringSliceDiffForDesiredState(tc.current, tc.desired, tc.desiredState)
			if !reflect.DeepEqual(gotAdd, tc.wantAdd) {
				t.Fatalf("computeStringSliceDiffForDesiredState() add mismatch: got=%v want=%v", gotAdd, tc.wantAdd)
			}
			if !reflect.DeepEqual(gotDel, tc.wantDel) {
				t.Fatalf("computeStringSliceDiffForDesiredState() remove mismatch: got=%v want=%v", gotDel, tc.wantDel)
			}
		})
	}
}
