package pkg

import (
	"bytes"
	"testing"
)

const TestString = "hello"

func TestRemoveN(t *testing.T) {
	validateRemoveX(t, "\n")
}

func TestRemoveRN(t *testing.T) {
	validateRemoveX(t, "\r\n")
}

func TestDontRemoveValid(t *testing.T) {
	validateRemoveX(t, "")
}

func validateRemoveX(t *testing.T, suffix string) {
	removed := RemoveCRLF([]byte(TestString + suffix))

	if bytes.Compare(removed, []byte(TestString)) != 0 {
		t.Error("\\n or \\r was not removed")
	}
}
