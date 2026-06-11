package auth

import (
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

func TestClientIDEnvOverridesBuiltin(t *testing.T) {
	t.Setenv("GITTY_CLIENT_ID", "from-env")
	if got := ClientID(); got != "from-env" {
		t.Fatalf("ClientID() = %q", got)
	}
}
