package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vromero/gofortune/pkg/fortune"
)

// TestPrepareRequestNoArgs verifies that with no positional arguments, the
// request is populated with the default fortune/offensive paths.
func TestPrepareRequestNoArgs(t *testing.T) {
	tmpDir := t.TempDir()
	fortuneDir := filepath.Join(tmpDir, "fortunes")
	offDir := filepath.Join(fortuneDir, "off")
	if err := os.MkdirAll(offDir, 0755); err != nil {
		t.Fatalf("failed to create fortune dirs: %v", err)
	}

	req, err := fortune.PrepareRequest(nil, fortuneDir, offDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(req.Paths) != 1 || req.Paths[0].Path != fortuneDir {
		t.Fatalf("expected default fortune path %q, got %+v", fortuneDir, req.Paths)
	}
	if len(req.OffensivePaths) != 1 || req.OffensivePaths[0].Path != offDir {
		t.Fatalf("expected default offensive path %q, got %+v", offDir, req.OffensivePaths)
	}
}

// TestPrepareRequestWithPath verifies that a positional path argument is
// captured as a ProbabilityPath with the expected Path value.
func TestPrepareRequestWithPath(t *testing.T) {
	args := []string{"/tmp/some/fortunes"}
	req, err := fortune.PrepareRequest(args, "/unused", "/unused/off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(req.Paths) != 1 {
		t.Fatalf("expected one path, got %d", len(req.Paths))
	}
	if req.Paths[0].Path != args[0] {
		t.Errorf("expected path %q, got %q", args[0], req.Paths[0].Path)
	}
	if req.Paths[0].Percentage != 0 {
		t.Errorf("expected default percentage 0, got %v", req.Paths[0].Percentage)
	}
}

// TestPrepareRequestWithPercentage verifies that a "N%" argument before a
// path sets that path's Percentage.
func TestPrepareRequestWithPercentage(t *testing.T) {
	args := []string{"30%", "/a", "70%", "/b"}
	req, err := fortune.PrepareRequest(args, "/unused", "/unused/off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(req.Paths) != 2 {
		t.Fatalf("expected two paths, got %d", len(req.Paths))
	}
	if req.Paths[0].Path != "/a" || req.Paths[0].Percentage != 30 {
		t.Errorf("path 0: expected (/a,30), got (%s,%v)", req.Paths[0].Path, req.Paths[0].Percentage)
	}
	if req.Paths[1].Path != "/b" || req.Paths[1].Percentage != 70 {
		t.Errorf("path 1: expected (/b,70), got (%s,%v)", req.Paths[1].Path, req.Paths[1].Percentage)
	}
}

// TestRootCmdFlagsRegistered verifies that all user-facing flags are wired up
// on RootCmd. Regression guard against flag-registration being dropped during
// refactors.
func TestRootCmdFlagsRegistered(t *testing.T) {
	expected := []string{
		"allMaxims", "offensive", "showCookieFile", "printListOfFiles",
		"considerAllEqual", "match", "longestShort", "longDictumsOnly",
		"shortOnly", "ignoreCase", "wait",
	}
	for _, name := range expected {
		if f := RootCmd.Flags().Lookup(name); f == nil {
			t.Errorf("flag %q is not registered on RootCmd", name)
		}
	}
}

// TestFortuneRunRejectsMissingPath verifies that fortuneRun surfaces an error
// when asked to load a path that doesn't exist, instead of panicking.
func TestFortuneRunRejectsMissingPath(t *testing.T) {
	req := fortune.Request{
		Paths:        []fortune.ProbabilityPath{{Path: filepath.Join(t.TempDir(), "does-not-exist")}},
		LongestShort: 160,
	}
	if err := fortuneRun(req); err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
}
