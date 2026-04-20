package fortune

import (
	"testing"
)

func TestSetProbabilitiesEqualSize(t *testing.T) {
	tests := []struct {
		name          string
		root          FileSystemNodeDescriptor
		expectPercent float32
	}{
		{
			name: "single file",
			root: FileSystemNodeDescriptor{
				NumFiles: 1,
				Children: []FileSystemNodeDescriptor{
					{NumEntries: 1, NumFiles: 1},
				},
			},
			expectPercent: 100,
		},
		{
			name: "two files equal size",
			root: FileSystemNodeDescriptor{
				NumFiles: 2,
				Children: []FileSystemNodeDescriptor{
					{NumEntries: 1, NumFiles: 2},
					{NumEntries: 1, NumFiles: 2},
				},
			},
			expectPercent: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetProbabilities(&tt.root, true)
			if tt.root.Percent != tt.expectPercent {
				t.Errorf("expected %f got %f", tt.expectPercent, tt.root.Percent)
			}
		})
	}
}

func TestCalculateUndefinedProbability_Empty(t *testing.T) {
	root := FileSystemNodeDescriptor{
		Children: []FileSystemNodeDescriptor{},
	}
	calculateUndefinedProbability(&root)
	if root.UndefinedChildrenPercent != 100 {
		t.Errorf("expected 100 got %f", root.UndefinedChildrenPercent)
	}
	if root.UndefinedNumEntries != 0 {
		t.Errorf("expected 0 got %d", root.UndefinedNumEntries)
	}
}

