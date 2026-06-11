package credential

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func lookupOK(account string) (string, error) {
	if account == "work" {
		return "gho_secret", nil
	}
	return "", errors.New("no token")
}

func TestGetPrintsUsernameAndToken(t *testing.T) {
	in := strings.NewReader("protocol=https\nhost=github.com\n\n")
	var out bytes.Buffer
	if err := Run("work", "get", in, &out, lookupOK); err != nil {
		t.Fatal(err)
	}
	want := "username=x-access-token\npassword=gho_secret\n\n"
	if out.String() != want {
		t.Fatalf("out = %q, want %q", out.String(), want)
	}
}

func TestGetUnknownAccountFails(t *testing.T) {
	var out bytes.Buffer
	if err := Run("ghost", "get", strings.NewReader(""), &out, lookupOK); err == nil {
		t.Fatal("expected error")
	}
}

func TestStoreAndEraseAreNoops(t *testing.T) {
	for _, action := range []string{"store", "erase"} {
		var out bytes.Buffer
		if err := Run("work", action, strings.NewReader("x=y\n"), &out, lookupOK); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if out.Len() != 0 {
			t.Fatalf("%s: unexpected output %q", action, out.String())
		}
	}
}
