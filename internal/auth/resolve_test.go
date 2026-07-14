package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stubProbes(t *testing.T, secretAlive, passReady bool) {
	t.Helper()
	origSecret, origPass := probeSecretService, passInitialized
	probeSecretService = func() bool { return secretAlive }
	passInitialized = func() bool { return passReady }
	t.Cleanup(func() {
		probeSecretService = origSecret
		passInitialized = origPass
	})
}

func TestResolveDarwinAlwaysSecretService(t *testing.T) {
	stubProbes(t, false, false)
	b, err := resolveBackendFor("darwin", "")
	if err != nil {
		t.Fatal(err)
	}
	if b.name() != "secret-service" {
		t.Fatalf("backend = %q", b.name())
	}
}

func TestResolveLinuxPrefersLiveSecretService(t *testing.T) {
	stubProbes(t, true, true)
	b, err := resolveBackendFor("linux", "")
	if err != nil {
		t.Fatal(err)
	}
	if b.name() != "secret-service" {
		t.Fatalf("backend = %q", b.name())
	}
}

func TestResolveLinuxFallsBackToPass(t *testing.T) {
	stubProbes(t, false, true)
	b, err := resolveBackendFor("linux", "")
	if err != nil {
		t.Fatal(err)
	}
	if b.name() != "pass" {
		t.Fatalf("backend = %q", b.name())
	}
}

func TestResolveLinuxNoBackendIsAGuidedError(t *testing.T) {
	stubProbes(t, false, false)
	_, err := resolveBackendFor("linux", "")
	if err == nil {
		t.Fatal("expected an error when no backend is available")
	}
	if !strings.Contains(err.Error(), "pass init") {
		t.Fatalf("error must guide the user to set up pass, got %q", err)
	}
}

func TestResolveOverrideForcesPass(t *testing.T) {
	stubProbes(t, true, false) // secret service alive, pass not initialised
	b, err := resolveBackendFor("linux", "pass")
	if err != nil {
		t.Fatal(err)
	}
	if b.name() != "pass" {
		t.Fatalf("backend = %q", b.name())
	}
}

func TestResolveOverrideForcesSecretService(t *testing.T) {
	stubProbes(t, false, true) // secret service dead, pass ready
	b, err := resolveBackendFor("linux", "secret-service")
	if err != nil {
		t.Fatal(err)
	}
	if b.name() != "secret-service" {
		t.Fatalf("backend = %q", b.name())
	}
}

func TestResolveUnknownOverrideErrors(t *testing.T) {
	stubProbes(t, true, true)
	if _, err := resolveBackendFor("linux", "bogus"); err == nil {
		t.Fatal("expected an error for an unknown backend override")
	}
}

func stubLookPath(t *testing.T, fn func(string) (string, error)) {
	t.Helper()
	orig := lookPath
	lookPath = fn
	t.Cleanup(func() {
		lookPath = orig
	})
}

func TestLivePassInitialized(t *testing.T) {
	t.Run("pass not on PATH", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".gpg-id"), []byte("key-id\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PASSWORD_STORE_DIR", dir)
		stubLookPath(t, func(string) (string, error) {
			return "", errors.New("not found")
		})
		if livePassInitialized() {
			t.Fatal("expected false when pass is not on PATH, even with .gpg-id present")
		}
	})

	t.Run("pass on PATH but no .gpg-id", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("PASSWORD_STORE_DIR", dir)
		stubLookPath(t, func(string) (string, error) {
			return "/usr/bin/pass", nil
		})
		if livePassInitialized() {
			t.Fatal("expected false when no .gpg-id is present in the store dir")
		}
	})

	t.Run("pass on PATH and .gpg-id present", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".gpg-id"), []byte("key-id\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PASSWORD_STORE_DIR", dir)
		stubLookPath(t, func(string) (string, error) {
			return "/usr/bin/pass", nil
		})
		if !livePassInitialized() {
			t.Fatal("expected true when pass is on PATH and .gpg-id is present")
		}
	})
}

func TestSecretServiceProbeAccountHasNoNUL(t *testing.T) {
	if strings.ContainsRune(secretServiceProbeAccount, '\x00') {
		t.Fatalf("secretServiceProbeAccount must not contain a NUL byte: %q", secretServiceProbeAccount)
	}
}
