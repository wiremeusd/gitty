package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/wiremeusd/gitty/internal/auth"
	"github.com/wiremeusd/gitty/internal/clone"
	"github.com/wiremeusd/gitty/internal/config"
	"github.com/wiremeusd/gitty/internal/github"
	"github.com/wiremeusd/gitty/internal/tui"
)

func runInteractive(cmd *cobra.Command, args []string) error {
	path, err := config.DefaultPath()
	if err != nil {
		return err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}
	if len(cfg.Accounts) == 0 {
		return errors.New("no accounts yet — run: gitty auth login")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	account, ok := cfg.AccountFor(cwd)
	if !ok {
		account, err = pickAccount(os.Stdin, os.Stdout, cfg.Accounts)
		if err != nil {
			return err
		}
		if askYesNo(os.Stdin, os.Stdout, "Remember this account for the current folder?") {
			cfg.Bind(cwd, account)
			if err := cfg.Save(path); err != nil {
				return err
			}
		}
	}

	token, err := auth.Token(account)
	if err != nil {
		return fmt.Errorf("no token for %s — run: gitty auth login (%w)", account, err)
	}

	cache, err := github.DefaultCache()
	if err != nil {
		return err
	}
	cached, _ := cache.Load(account) // a corrupt cache is not a reason to fail
	client := github.NewClient(token)
	ctx := cmd.Context()
	fetch := func() ([]github.Repo, error) {
		repos, err := client.Repos(ctx)
		if err != nil {
			return nil, err
		}
		_ = cache.Save(account, repos)
		return repos, nil
	}

	// No cache — load synchronously so that 401 and network errors
	// are shown to the user as text rather than an empty list.
	if len(cached) == 0 {
		cached, err = fetch()
		if errors.Is(err, github.ErrUnauthorized) {
			return fmt.Errorf("token for account %s has been revoked — run: gitty auth login", account)
		}
		if err != nil {
			return fmt.Errorf("failed to fetch repository list: %w", err)
		}
		fetch = nil // data is already fresh, no background refresh needed
		if len(cached) == 0 {
			fmt.Printf("Account %s has no accessible repositories.\n", account)
			return nil
		}
	}

	final, err := tea.NewProgram(tui.New(account, cached, fetch), tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}
	sel := final.(tui.Model).Selected()
	if sel == nil {
		return nil // exited without a selection — not an error
	}

	fmt.Printf("Cloning %s…\n", sel.FullName)
	if err := clone.Run(*sel, account, cwd, clone.GitRunner); err != nil {
		return err
	}
	fmt.Printf("✔ Cloned into ./%s\n", sel.Name)
	return nil
}
