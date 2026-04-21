package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vromero/gofortune/pkg/fortune"
)

func TestFortuneRun(t *testing.T) {
	// Create a temporary directory for our test fortune files
	tmpDir, err := os.MkdirTemp("", "fortunes_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fortuneDir := filepath.Join(tmpDir, "fortunes")
	offDir := filepath.Join(fortuneDir, "off")
	err = os.MkdirAll(offDir, 0755)
	if err != nil {
		t.Fatalf("failed to create fortune dirs: %v", err)
	}

	// Create a dummy fortune file
	fortuneFile := filepath.Join(fortuneDir, "test_fortune")
	content := []byte("Fortune 1\nFortune 2\n")
	err = os.WriteFile(fortuneFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to write fortune file: %v", err)
	}

	// This test is very skeletal and just checks if we can execute the command logic
	// with some dummy paths.
	args := []string{fortuneDir}
	request := fortune.PrepareRequest(args, fortuneDir, offDir)
	
	if len(request.Paths) == 0 {
		t.Errorf("expected at least one path in request")
	}
	if request.Paths[0].Path != fortuneDir {
		t.Errorf("expected path %s, got %s", fortuneDir, request.Paths[0].Path)
	}
}
