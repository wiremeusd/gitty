package auth

import "github.com/zalando/go-keyring"

// Tokens are stored in the macOS Keychain: service=gitty, account=GitHub account name.
const keyringService = "gitty"

func SaveToken(account, token string) error {
	return keyring.Set(keyringService, account, token)
}

func Token(account string) (string, error) {
	return keyring.Get(keyringService, account)
}

func DeleteToken(account string) error {
	return keyring.Delete(keyringService, account)
}
