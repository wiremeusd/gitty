package auth

import (
	"os"

	"github.com/zalando/go-keyring"
)

// Tokens are stored under service=gitty, account=GitHub account name.
const keyringService = "gitty"

// ErrNotFound is returned by Token when no token exists for the account.
// Every backend maps its own "missing" outcome to this sentinel.
var ErrNotFound = keyring.ErrNotFound

// backend is a token storage backend. secretService is the only
// implementation today; later tasks add more and a resolver that picks
// between them.
type backend interface {
	set(account, token string) error
	get(account string) (string, error)
	delete(account string) error
	name() string
}

func SaveToken(account, token string) error {
	return secretService{}.set(account, token)
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
	return secretService{}.get(account)
}

func DeleteToken(account string) error {
	return secretService{}.delete(account)
}
