package auth

import "github.com/zalando/go-keyring"

// secretService stores tokens in the system keyring via D-Bus Secret Service
// (Linux) or the login Keychain (macOS): service=gitty, account=GitHub name.
type secretService struct{}

var _ backend = secretService{}

func (secretService) name() string { return "secret-service" }

func (secretService) set(account, token string) error {
	return keyring.Set(keyringService, account, token)
}

func (secretService) get(account string) (string, error) {
	return keyring.Get(keyringService, account)
}

func (secretService) delete(account string) error {
	return keyring.Delete(keyringService, account)
}
