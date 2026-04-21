package strfile

import (
	"bytes"
	"testing"

	"github.com/vromero/gofortune/pkg"
)

func TestAdvanceAwareSplitterEmptyAtEOF(t *testing.T) {
	advance, token, err := advanceAwareSplitter(nil, true)
	if advance != 0 || token != nil || err != nil {
		t.Fatalf("expected (0,nil,nil), got (%d,%v,%v)", advance, token, err)
	}
}

func TestAdvanceAwareSplitterCompleteLine(t *testing.T) {
	data := []byte("hello\nworld\n")
	advance, token, err := advanceAwareSplitter(data, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if advance != 6 {
		t.Errorf("expected advance=6, got %d", advance)
	}
	if !bytes.Equal(token, []byte("hello\n")) {
		t.Errorf("expected token %q, got %q", "hello\n", token)
	}
}

func TestAdvanceAwareSplitterRequestMoreData(t *testing.T) {
	// No newline and not at EOF: splitter should ask for more data.
	advance, token, err := advanceAwareSplitter([]byte("partial"), false)
	if advance != 0 || token != nil || err != nil {
		t.Fatalf("expected (0,nil,nil), got (%d,%v,%v)", advance, token, err)
	}
}

func TestAdvanceAwareSplitterFinalUnterminatedLine(t *testing.T) {
	data := []byte("tail")
	advance, token, err := advanceAwareSplitter(data, true)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if advance != len(data) {
		t.Errorf("expected advance=%d, got %d", len(data), advance)
	}
	if !bytes.Equal(token, data) {
		t.Errorf("expected token %q, got %q", data, token)
	}
}

// TestShufflePreservesElements verifies that Shuffle is a permutation: same
// length, same multiset of elements (regardless of order).
func TestShufflePreservesElements(t *testing.T) {
	original := []pkg.DataPos{
		{OriginalOffset: 0},
		{OriginalOffset: 1},
		{OriginalOffset: 2},
		{OriginalOffset: 3},
		{OriginalOffset: 4},
	}
	shuffled := make([]pkg.DataPos, len(original))
	copy(shuffled, original)

	Shuffle(shuffled)

	if len(shuffled) != len(original) {
		t.Fatalf("length changed: before=%d after=%d", len(original), len(shuffled))
	}

	seen := make(map[uint32]int, len(original))
	for _, d := range shuffled {
		seen[d.OriginalOffset]++
	}
	for _, d := range original {
		if seen[d.OriginalOffset] != 1 {
			t.Errorf("offset %d appeared %d times, expected 1", d.OriginalOffset, seen[d.OriginalOffset])
		}
	}
}

// TestShuffleEmpty is a regression guard: Shuffle on empty input must not panic.
func TestShuffleEmpty(t *testing.T) {
	Shuffle(nil)
	Shuffle([]pkg.DataPos{})
}
