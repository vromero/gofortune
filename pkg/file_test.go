package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveFileExtension(t *testing.T) {
	got := RemoveFileExtension("hello.goodbye")
	expected := "hello"

	if got != expected {
		t.Error("Expected " + expected + " got " + got)
	}
}

func TestFileExists(t *testing.T) {
	existentFile, err := os.CreateTemp("", "gofortune")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = existentFile.Close()
		_ = os.Remove(existentFile.Name())
	})

	if !FileExists(existentFile.Name()) {
		t.Error("expected file to exist")
	}
}

func TestFileDoestExists(t *testing.T) {
	emptyDirectory := t.TempDir()
	nonExistentFile := filepath.Join(emptyDirectory, "nonExistentFile.name")

	if FileExists(nonExistentFile) {
		t.Error("expected file to not exist")
	}
}
