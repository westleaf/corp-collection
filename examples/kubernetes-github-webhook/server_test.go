package main

import (
	"testing"

	"github.com/google/go-github/v69/github"
)

func TestGetFiles(t *testing.T) {
	files := getFiles([]*github.HeadCommit{{Added: []string{"file1"}, Modified: []string{"file2"}}})
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}
