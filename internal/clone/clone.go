// Package clone runs the system git clone over HTTPS and registers
// gitty as the credential helper in the new repository's config.
package clone

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wiremeusd/gitty/internal/github"
)

type Runner func(name string, args ...string) error

// GitRunner is the production Runner: it pipes git output to the terminal
// so that clone progress is visible.
func GitRunner(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run clones repo into parentDir/<repository name>.
// The first "-c credential.helper=" clears any global helpers (e.g.
// osxkeychain) so credentials from a different account are not used; the second
// registers gitty. "git clone -c" writes both into the new repository's config,
// so subsequent pull/push operations also go through gitty.
func Run(repo github.Repo, account, parentDir string, run Runner) error {
	// Empty name check — before main validation.
	if account == "" {
		return errors.New("empty account name")
	}
	// The helper string ends up in .git/config and is executed by git via sh -c,
	// so the account name is strictly validated (GitHub logins: [A-Za-z0-9-]).
	for _, r := range account {
		ok := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-'
		if !ok {
			return fmt.Errorf("invalid account name: %q", account)
		}
	}

	// Pre-check for a clear error message; if the directory appears between
	// this check and the git invocation, git will refuse to overwrite it anyway.
	dest := filepath.Join(parentDir, repo.Name)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("directory %s already exists — clone cancelled", dest)
	}
	helper := "credential.helper=!gitty credential --account " + account
	if err := run("git", "clone",
		"-c", "credential.helper=",
		"-c", helper,
		repo.CloneURL, dest); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}
