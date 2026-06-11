package auth

import "os"

// builtinClientID is the public Client ID of the GitHub OAuth application.
// It is injected at release build time:
//
//	go build -ldflags "-X github.com/wiremeusd/gitty/internal/auth.builtinClientID=<id>"
var builtinClientID = ""

// ClientID: the GITTY_CLIENT_ID environment variable takes priority
// (useful for development before the OAuth application is registered).
func ClientID() string {
	if v := os.Getenv("GITTY_CLIENT_ID"); v != "" {
		return v
	}
	return builtinClientID
}
