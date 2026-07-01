package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestTokenRoundtrip(t *testing.T) {
	keyring.MockInit() // in-memory keyring, does not touch the real Keychain

	if err := SaveToken("work", "gho_abc"); err != nil {
		t.Fatal(err)
	}
	token, err := Token("work")
	if err != nil {
		t.Fatal(err)
	}
	if token != "gho_abc" {
		t.Fatalf("token = %q", token)
	}
	if err := DeleteToken("work"); err != nil {
		t.Fatal(err)
	}
	if _, err := Token("work"); err == nil {
		t.Fatal("expected error after delete")
	}
}

// On a headless server the Secret Service keyring is unavailable, so the token
// must be readable from the GITTY_TOKEN environment variable instead — set it in
// a shell profile or a systemd Environment= line, no keyring daemon required.
func TestTokenEnvOverridesKeyring(t *testing.T) {
	keyring.MockInit() // empty in-memory keyring — nothing stored for "work"
	t.Setenv("GITTY_TOKEN", "gho_from_env")
	token, err := Token("work")
	if err != nil {
		t.Fatalf("Token with GITTY_TOKEN set must not error: %v", err)
	}
	if token != "gho_from_env" {
		t.Fatalf("token = %q, want the env value", token)
	}
}

// The env var wins even when a different token is stored in the keyring, so an
// operator can override without clearing the keyring.
func TestTokenEnvTakesPriorityOverStored(t *testing.T) {
	keyring.MockInit()
	if err := SaveToken("work", "gho_keyring"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITTY_TOKEN", "gho_env")
	token, err := Token("work")
	if err != nil {
		t.Fatal(err)
	}
	if token != "gho_env" {
		t.Fatalf("token = %q, want the env value to win", token)
	}
}

// A blank GITTY_TOKEN is treated as unset, so it can't shadow the keyring.
func TestTokenBlankEnvFallsBackToKeyring(t *testing.T) {
	keyring.MockInit()
	if err := SaveToken("work", "gho_keyring"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITTY_TOKEN", "")
	token, err := Token("work")
	if err != nil {
		t.Fatal(err)
	}
	if token != "gho_keyring" {
		t.Fatalf("token = %q, want the keyring value when env is blank", token)
	}
}

func TestLinuxHintWrapsKeyringError(t *testing.T) {
	base := errors.New("The name org.freedesktop.secrets was not provided by any .service files")
	err := withLinuxHint("linux", base)
	if !errors.Is(err, base) {
		t.Fatalf("wrapped error must match the original via errors.Is, got %v", err)
	}
	if !strings.Contains(err.Error(), "Secret Service") {
		t.Fatalf("expected a Secret Service hint, got %q", err)
	}
}

func TestLinuxHintLeavesNotFoundAlone(t *testing.T) {
	if err := withLinuxHint("linux", keyring.ErrNotFound); err != keyring.ErrNotFound {
		t.Fatalf("ErrNotFound must pass through unchanged, got %v", err)
	}
}

func TestLinuxHintNilStaysNil(t *testing.T) {
	if err := withLinuxHint("linux", nil); err != nil {
		t.Fatalf("nil must stay nil, got %v", err)
	}
}

func TestLinuxHintSkippedOnOtherOS(t *testing.T) {
	base := errors.New("keychain denied")
	if err := withLinuxHint("darwin", base); err != base {
		t.Fatalf("non-linux error must pass through unchanged, got %v", err)
	}
}

func TestClientIDEnvOverridesBuiltin(t *testing.T) {
	t.Setenv("GITTY_CLIENT_ID", "from-env")
	if got := ClientID(); got != "from-env" {
		t.Fatalf("ClientID() = %q", got)
	}
}
