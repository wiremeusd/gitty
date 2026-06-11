package github

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheMissReturnsNilWithoutError(t *testing.T) {
	c := &Cache{Dir: t.TempDir()}
	repos, err := c.Load("nobody")
	if err != nil {
		t.Fatal(err)
	}
	if repos != nil {
		t.Fatalf("repos = %v, want nil", repos)
	}
}

func TestCacheRoundtrip(t *testing.T) {
	c := &Cache{Dir: t.TempDir() + "/nested"} // Save must create the directory
	in := []Repo{{FullName: "me/one", Name: "one", Stars: 5}}
	if err := c.Save("me", in); err != nil {
		t.Fatal(err)
	}
	out, err := c.Load("me")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].FullName != "me/one" || out[0].Stars != 5 {
		t.Fatalf("out = %+v", out)
	}
}

func TestCacheCorruptionIsMiss(t *testing.T) {
	dir := t.TempDir()
	c := &Cache{Dir: dir}
	if err := os.WriteFile(filepath.Join(dir, "me.json"), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	repos, err := c.Load("me")
	if err != nil {
		t.Fatalf("corrupt cache must not return error, got: %v", err)
	}
	if repos != nil {
		t.Fatal("corrupt cache must return nil repos")
	}
}

func TestCacheRejectsUnsafeAccountNames(t *testing.T) {
	c := &Cache{Dir: t.TempDir()}
	if err := c.Save("../evil", []Repo{{Name: "x"}}); err == nil {
		t.Fatal("Save must reject account names with path separators")
	}
	if repos, err := c.Load("../evil"); err == nil || repos != nil {
		t.Fatalf("Load must reject unsafe names, got repos=%v err=%v", repos, err)
	}
}
