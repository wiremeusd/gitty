package auth

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/zalando/go-keyring"
)

// Tokens are stored in the system keyring (macOS Keychain / Linux Secret Service):
// service=gitty, account=GitHub account name.
const keyringService = "gitty"

func SaveToken(account, token string) error {
	return withLinuxHint(runtime.GOOS, keyring.Set(keyringService, account, token))
}

// Token returns the GitHub token for account. The GITTY_TOKEN environment
// variable takes priority over the keyring: on headless servers (no Secret
// Service daemon) an operator exports a PAT there — via a shell profile or a
// systemd Environment= line — instead of running `gitty auth login`. It applies
// to every account, so a single-purpose box needs no keyring at all.
func Token(account string) (string, error) {
	if v := os.Getenv("GITTY_TOKEN"); v != "" {
		return v, nil
	}
	token, err := keyring.Get(keyringService, account)
	return token, withLinuxHint(runtime.GOOS, err)
}

func DeleteToken(account string) error {
	return withLinuxHint(runtime.GOOS, keyring.Delete(keyringService, account))
}

// On Linux the keyring is the D-Bus Secret Service; when no daemon is running
// (headless server, WSL, minimal WM) go-keyring's raw D-Bus error is cryptic,
// so point the user at the fix. "Not found" is a normal outcome, not a setup issue.
func withLinuxHint(goos string, err error) error {
	if err == nil || errors.Is(err, keyring.ErrNotFound) || goos != "linux" {
		return err
	}
	return fmt.Errorf("%w\nhint: gitty stores tokens in the Secret Service keyring; make sure a keyring daemon is running (GNOME Keyring or KWallet) — on headless systems install and unlock gnome-keyring first", err)
}
