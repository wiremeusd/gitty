package clone

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wiremeusd/gitty/internal/github"
)

func TestRunBuildsGitCloneCommand(t *testing.T) {
	dir := t.TempDir()
	var gotName string
	var gotArgs []string
	fake := func(name string, args ...string) error {
		gotName, gotArgs = name, args
		return nil
	}

	repo := github.Repo{Name: "proj", FullName: "me/proj", CloneURL: "https://github.com/me/proj.git"}
	if err := Run(repo, "work", dir, fake); err != nil {
		t.Fatal(err)
	}
	if gotName != "git" {
		t.Fatalf("name = %q", gotName)
	}
	want := []string{
		"clone",
		"-c", "credential.helper=",
		"-c", "credential.helper=!gitty credential --account work",
		"https://github.com/me/proj.git",
		filepath.Join(dir, "proj"),
	}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("args = %v\nwant %v", gotArgs, want)
	}
}

func TestRunRefusesExistingDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "proj"), 0o755); err != nil {
		t.Fatal(err)
	}
	called := false
	fake := func(string, ...string) error { called = true; return nil }

	repo := github.Repo{Name: "proj", CloneURL: "https://x"}
	if err := Run(repo, "work", dir, fake); err == nil {
		t.Fatal("expected error for existing dir")
	}
	if called {
		t.Fatal("git must not be invoked")
	}
}

func TestRunRejectsUnsafeAccountName(t *testing.T) {
	dir := t.TempDir()
	called := false
	fake := func(string, ...string) error { called = true; return nil }

	repo := github.Repo{Name: "proj", CloneURL: "https://x"}
	if err := Run(repo, `evil"; rm -rf ~; "`, dir, fake); err == nil {
		t.Fatal("expected error for unsafe account name")
	}
	if called {
		t.Fatal("git must not be invoked")
	}
}
