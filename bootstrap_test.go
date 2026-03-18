package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureJSONFileCreatesFormattedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.json")

	if err := ensureJSONFile(path, map[string]string{"hello": "world"}); err != nil {
		t.Fatalf("ensureJSONFile returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	want := "{\n  \"hello\": \"world\"\n}\n"
	if string(data) != want {
		t.Fatalf("unexpected JSON content:\nwant: %q\ngot:  %q", want, string(data))
	}
}
