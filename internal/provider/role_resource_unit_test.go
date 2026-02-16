package provider

import (
	"reflect"
	"testing"
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
