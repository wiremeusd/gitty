// Package credential implements the git credential helper protocol.
// gitty is registered in the cloned repository's config as
//
//	credential.helper = "!gitty credential --account <name>"
//
// and git appends the action (get/store/erase) to that string.
package credential

import (
	"fmt"
	"io"
)

// Run handles a single protocol action. lookup retrieves a token by account name
// (auth.Token in production, a stub in tests).
func Run(account, action string, in io.Reader, out io.Writer, lookup func(string) (string, error)) error {
	// The protocol requires reading stdin to completion regardless of the action.
	if _, err := io.ReadAll(in); err != nil {
		return err
	}
	if action != "get" {
		// store/erase (and capability in newer git versions) are intentionally ignored:
		// tokens are managed by gitty itself, git does not store them.
		return nil
	}
	token, err := lookup(account)
	if err != nil {
		return fmt.Errorf("no token for account %q: %w", account, err)
	}
	// The trailing empty line is the attribute-list terminator in the git protocol.
	_, err = fmt.Fprintf(out, "username=x-access-token\npassword=%s\n\n", token)
	return err
}
