package glmodules

import (
	"encoding/base64"
	"testing"
)

func TestTextToBase64(t *testing.T) {
	text := "Running and Walking"
	expected := base64.StdEncoding.EncodeToString([]byte(text))

	if len(expected) == 0 {
		t.Errorf("empty encoded string")
	}
	print(expected)
}

func TestBase64ToText(t *testing.T) {
	base64Str := "UnVubmluZyBhbmQgV2Fsa2luZw=="

	decodedBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		t.Fatalf("Failed to decode base64 string: %v", err)
	}

	result := string(decodedBytes)

	if len(result) == 0 {
		t.Errorf("Empty decoded base64")
	}
	print(result)
}
