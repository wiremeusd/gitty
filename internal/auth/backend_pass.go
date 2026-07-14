package auth

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// passStore stores tokens in a `pass` password store as gitty/<account>.
// GPG encryption and passphrase caching are handled entirely by pass/gpg-agent,
// so on a headless VPS the token is encrypted at rest and unlocked once per
// session (like the macOS Keychain).
type passStore struct {
	run func(args []string, stdin string) (string, error)
}

func newPassStore() *passStore {
	return &passStore{run: runPass}
}

func (passStore) name() string { return "pass" }

func passEntry(account string) string { return "gitty/" + account }

func (p *passStore) set(account, token string) error {
	// -e reads a single line from stdin; -f overwrites without prompting.
	_, err := p.run([]string{"insert", "-e", "-f", passEntry(account)}, token+"\n")
	return err
}

func (p *passStore) get(account string) (string, error) {
	out, err := p.run([]string{"show", passEntry(account)}, "")
	if err != nil {
		if isPassNotFound(err) {
			return "", ErrNotFound
		}
		return "", err
	}
	// The token is the first line; pass may print extra metadata lines after it.
	if i := strings.IndexByte(out, '\n'); i >= 0 {
		out = out[:i]
	}
	return strings.TrimRight(out, "\r"), nil
}

func (p *passStore) delete(account string) error {
	_, err := p.run([]string{"rm", "-f", passEntry(account)}, "")
	if err != nil && isPassNotFound(err) {
		return nil // deleting a missing entry is a no-op, matching keyring.Delete semantics
	}
	return err
}

func isPassNotFound(err error) bool {
	return strings.Contains(err.Error(), "is not in the password store")
}

func runPass(args []string, stdin string) (string, error) {
	cmd := exec.Command("pass", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(errb.String()); msg != "" {
			return out.String(), fmt.Errorf("pass %s: %s: %w", args[0], msg, err)
		}
		return out.String(), fmt.Errorf("pass %s: %w", args[0], err)
	}
	return out.String(), nil
}
