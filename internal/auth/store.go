package auth

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/zalando/go-keyring"
)

// Tokens are stored in the system keyring (macOS Keychain / Linux Secret Service):
// service=gitty, account=GitHub account name.
const keyringService = "gitty"

func SaveToken(account, token string) error {
	return withLinuxHint(runtime.GOOS, keyring.Set(keyringService, account, token))
}

func Token(account string) (string, error) {
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
