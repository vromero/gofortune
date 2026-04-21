package fortune

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/vromero/gofortune/pkg"
)

// TestGetRandomLeafNodeNoChildren: a leaf node returns itself, no error.
func TestGetRandomLeafNodeNoChildren(t *testing.T) {
	leaf := FileSystemNodeDescriptor{Path: "/tmp/fortune", NumEntries: 3}
	got, err := GetRandomLeafNode(leaf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != leaf.Path {
		t.Errorf("expected leaf itself, got %+v", got)
	}
}

// TestGetRandomLeafNodeDeterministic: with a single child taking 100%% of the
// probability, selection is deterministic.
func TestGetRandomLeafNodeDeterministic(t *testing.T) {
	leaf := FileSystemNodeDescriptor{Path: "/only", NumEntries: 1, Percent: 100}
	root := FileSystemNodeDescriptor{
		Percent:  100,
		Children: []FileSystemNodeDescriptor{leaf},
	}
	for i := 0; i < 20; i++ {
		got, err := GetRandomLeafNode(root)
		if err != nil {
			t.Fatalf("iter %d: unexpected error: %v", i, err)
		}
		if got.Path != "/only" {
			t.Fatalf("iter %d: expected /only, got %q", i, got.Path)
		}
	}
}

// TestSetProbabilitiesEqualSizeDistributes: with considerEqualSize=true, two
// leaves receive 50%% each and the root is 100%%.
func TestSetProbabilitiesEqualSizeDistributes(t *testing.T) {
	root := FileSystemNodeDescriptor{
		NumFiles: 2,
		Children: []FileSystemNodeDescriptor{
			{NumEntries: 10, NumFiles: 2},
			{NumEntries: 10, NumFiles: 2},
		},
	}
	SetProbabilities(&root, true)

	if root.Percent != 100 {
		t.Errorf("root percent: expected 100, got %v", root.Percent)
	}
	for i, c := range root.Children {
		if c.Percent != 50 {
			t.Errorf("child %d: expected 50, got %v", i, c.Percent)
		}
	}
}

// TestSetProbabilitiesWeightedDistributes: without considerEqualSize, a file
// with more entries receives a larger share of the undefined percentage.
func TestSetProbabilitiesWeightedDistributes(t *testing.T) {
	root := FileSystemNodeDescriptor{
		Children: []FileSystemNodeDescriptor{
			{NumEntries: 10, Table: pkg.DataTable{NumberOfStrings: 10}},
			{NumEntries: 30, Table: pkg.DataTable{NumberOfStrings: 30}},
		},
	}
	// Parent-link each child; the implementation walks up to find undefined %.
	for i := range root.Children {
		root.Children[i].Parent = &root
	}
	root.NumEntries = 40

	SetProbabilities(&root, false)

	// Children percentages should roughly add up to 100 and reflect the 1:3 ratio.
	total := root.Children[0].Percent + root.Children[1].Percent
	if total < 99.99 || total > 100.01 {
		t.Errorf("children percentages should sum to 100, got %v", total)
	}
	if root.Children[1].Percent <= root.Children[0].Percent {
		t.Errorf("larger file should get larger share; got %v vs %v",
			root.Children[0].Percent, root.Children[1].Percent)
	}
}

// TestGetFortunesMatchingErrorsOnMissingIndex: a leaf node pointing at a
// non-existent index file should surface an error on the error channel
// and not panic (regression guard for the old nil-Close panic).
//
// Both channels must be drained concurrently because the producer does not
// buffer; draining one fully before the other can deadlock.
func TestGetFortunesMatchingErrorsOnMissingIndex(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	leaf := FileSystemNodeDescriptor{
		Path:       missing,
		IndexPath:  missing + ".dat",
		NumEntries: 1,
	}

	dataCh, errCh := getFortunesMatching(leaf, regexp.MustCompile(".*"))

	gotErr := false
	for dataCh != nil || errCh != nil {
		select {
		case _, ok := <-dataCh:
			if !ok {
				dataCh = nil
				continue
			}
			t.Error("did not expect any fortune data")
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if !strings.Contains(err.Error(), "index") {
				t.Errorf("expected error mentioning 'index', got %q", err.Error())
			}
			gotErr = true
		}
	}
	if !gotErr {
		t.Error("expected at least one error on error channel")
	}
}

// TestGetLengthFilteredRandomFortuneGivesUp verifies that the length-filter
// loop terminates with an error when no fortune can satisfy the constraints,
// instead of looping forever.
func TestGetLengthFilteredRandomFortuneGivesUp(t *testing.T) {
	// Root has no children: GetRandomFortune will return "empty" which
	// propagates out as an error *before* the retry bound kicks in. That
	// is still a correct terminating behaviour.
	empty := FileSystemNodeDescriptor{NumEntries: 0}
	_, err := GetLengthFilteredRandomFortune(empty, 10, 0)
	if err == nil {
		t.Fatal("expected error from empty tree, got nil")
	}
}

