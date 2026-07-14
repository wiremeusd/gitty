package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wiremeusd/gitty/internal/auth"
	"github.com/wiremeusd/gitty/internal/config"
	"github.com/wiremeusd/gitty/internal/github"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage GitHub accounts",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthListCmd(), newAuthLogoutCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Add an account via OAuth Device Flow",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID := auth.ClientID()
			if clientID == "" {
				return errors.New("Client ID not set: set GITTY_CLIENT_ID or build a release binary (see README)")
			}
			ctx := cmd.Context()
			flow := auth.NewDeviceFlow(clientID)
			dc, err := flow.RequestCode(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("Open %s and enter the code: %s\n", dc.VerificationURI, dc.UserCode)
			fmt.Println("Waiting for confirmation in the browser…")
			token, err := flow.PollToken(ctx, dc)
			if err != nil {
				return err
			}
			login, err := github.NewClient(token).Login(ctx)
			if err != nil {
				return err
			}
			if err := auth.SaveToken(login, token); err != nil {
				return fmt.Errorf("could not save token: %w", err)
			}
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			cfg.AddAccount(login)
			if err := cfg.Save(path); err != nil {
				return err
			}
			fmt.Printf("✔ Account %s added\n", login)
			return nil
		},
	}
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List added accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			if len(cfg.Accounts) == 0 {
				fmt.Println("no accounts yet — run: gitty auth login")
				return nil
			}
			for _, a := range cfg.Accounts {
				fmt.Println(a)
			}
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout <account>",
		Short: "Remove an account: token from the system keyring and bindings from config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			if err := auth.DeleteToken(name); err != nil {
				fmt.Printf("Warning: token not found in the system keyring (%v)\n", err)
			}
			cfg.RemoveAccount(name)
			if err := cfg.Save(path); err != nil {
				return err
			}
			fmt.Printf("✔ Account %s removed\n", name)
			return nil
		},
	}
}
