package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsEmptyConfig(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Accounts) != 0 || len(cfg.Bindings) != 0 {
		t.Fatalf("expected empty config, got %+v", cfg)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.toml")
	cfg := &Config{Bindings: map[string]string{}}
	cfg.AddAccount("work")
	cfg.AddAccount("work") // duplicate is not added
	cfg.Bind("/Users/x/work", "work")
	if err := cfg.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got.Accounts) != 1 || got.Accounts[0] != "work" {
		t.Fatalf("accounts = %v", got.Accounts)
	}
	if got.Bindings["/Users/x/work"] != "work" {
		t.Fatalf("bindings = %v", got.Bindings)
	}
}

func TestAccountForDeepestBindingWins(t *testing.T) {
	cfg := &Config{
		Accounts: []string{"personal", "work", "client"},
		Bindings: map[string]string{
			"/Users/x/work":         "work",
			"/Users/x/work/clients": "client",
		},
	}
	acc, ok := cfg.AccountFor("/Users/x/work/clients/proj")
	if !ok || acc != "client" {
		t.Fatalf("got %q, %v; want client", acc, ok)
	}
	acc, ok = cfg.AccountFor("/Users/x/work/other")
	if !ok || acc != "work" {
		t.Fatalf("got %q, %v; want work", acc, ok)
	}
}

func TestAccountForDoesNotMatchPathPrefixOfSibling(t *testing.T) {
	// /Users/x/work must not match /Users/x/workspace
	cfg := &Config{
		Accounts: []string{"a", "b"},
		Bindings: map[string]string{"/Users/x/work": "a"},
	}
	if _, ok := cfg.AccountFor("/Users/x/workspace"); ok {
		t.Fatal("workspace must not match binding for work")
	}
}

func TestAccountForSingleAccountFallback(t *testing.T) {
	cfg := &Config{Accounts: []string{"solo"}, Bindings: map[string]string{}}
	acc, ok := cfg.AccountFor("/anywhere")
	if !ok || acc != "solo" {
		t.Fatalf("got %q, %v; want solo", acc, ok)
	}
}

func TestAccountForNoBindingMultipleAccounts(t *testing.T) {
	cfg := &Config{Accounts: []string{"a", "b"}, Bindings: map[string]string{}}
	if _, ok := cfg.AccountFor("/anywhere"); ok {
		t.Fatal("expected no account")
	}
}

func TestRemoveAccountDropsItsBindings(t *testing.T) {
	cfg := &Config{
		Accounts: []string{"a", "b"},
		Bindings: map[string]string{"/x": "a", "/y": "b"},
	}
	cfg.RemoveAccount("a")
	if len(cfg.Accounts) != 1 || cfg.Accounts[0] != "b" {
		t.Fatalf("accounts = %v", cfg.Accounts)
	}
	if _, ok := cfg.Bindings["/x"]; ok {
		t.Fatal("binding for removed account must be gone")
	}
	if cfg.Bindings["/y"] != "b" {
		t.Fatal("binding of other account must survive")
	}
}

func TestLoadNormalisesHandEditedBindingKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	raw := "accounts = [\"work\"]\n\n[bindings]\n\"/Users/x/work/\" = \"work\"\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	acc, ok := cfg.AccountFor("/Users/x/work/proj")
	if !ok || acc != "work" {
		t.Fatalf("got %q, %v; want work", acc, ok)
	}
}
