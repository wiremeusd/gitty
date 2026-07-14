package auth

import (
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
