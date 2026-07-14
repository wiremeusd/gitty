package auth

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/zalando/go-keyring"
)

// Overridable in tests.
var (
	probeSecretService = liveSecretService
	passInitialized    = livePassInitialized
	lookPath           = exec.LookPath
)

// secretServiceProbeAccount is a nonexistent account name used solely to
// probe whether a Secret Service daemon is answering. It must not contain a
// NUL byte: the D-Bus string encoder rejects NUL bytes with a marshal error
// rather than returning ErrNotFound, which would make a live daemon look
// dead.
const secretServiceProbeAccount = "gitty-secretservice-probe"

func resolveBackend() (backend, error) {
	return resolveBackendFor(runtime.GOOS, os.Getenv("GITTY_KEYRING_BACKEND"))
}

func resolveBackendFor(goos, override string) (backend, error) {
	switch override {
	case "secret-service":
		return secretService{}, nil
	case "pass":
		return newPassStore(), nil
	case "", "auto":
		// fall through to auto-detection
	default:
		return nil, fmt.Errorf("unknown GITTY_KEYRING_BACKEND %q (want auto, secret-service, or pass)", override)
	}

	if goos != "linux" {
		return secretService{}, nil
	}
	if probeSecretService() {
		return secretService{}, nil
	}
	if passInitialized() {
		return newPassStore(), nil
	}
	return nil, errNoBackend
}

var errNoBackend = errors.New(
	"no usable keyring: gitty needs a running Secret Service (desktop) or an initialised pass store (headless/VPS).\n" +
		"one-time pass setup: gpg --quick-gen-key \"you@example.com\" && pass init <key-id>\n" +
		"see: https://github.com/wiremeusd/gitty#running-on-a-vps--headless-server")

// liveSecretService reports whether a Secret Service daemon is answering.
// A "not found" reply means the daemon is up (it answered the query);
// any other error means no daemon is reachable.
func liveSecretService() bool {
	_, err := keyring.Get(keyringService, secretServiceProbeAccount)
	// Any error other than ErrNotFound (no daemon, locked/dismissed keyring) counts as unavailable; force with GITTY_KEYRING_BACKEND=secret-service.
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}

// livePassInitialized reports whether the pass binary exists and its store
// has been initialised (a .gpg-id file is present).
func livePassInitialized() bool {
	if _, err := lookPath("pass"); err != nil {
		return false
	}
	store := os.Getenv("PASSWORD_STORE_DIR")
	if store == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		store = filepath.Join(home, ".password-store")
	}
	_, err := os.Stat(filepath.Join(store, ".gpg-id"))
	return err == nil
}
