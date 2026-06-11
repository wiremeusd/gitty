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
