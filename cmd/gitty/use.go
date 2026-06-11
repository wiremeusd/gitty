package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/wiremeusd/gitty/internal/config"
)

func newUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <account>",
		Short: "Bind the current folder (and subfolders) to an account",
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
			if !slices.Contains(cfg.Accounts, name) {
				return fmt.Errorf("account %q not found — run first: gitty auth login", name)
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			cfg.Bind(cwd, name)
			if err := cfg.Save(path); err != nil {
				return err
			}
			fmt.Printf("✔ %s → %s\n", cwd, name)
			return nil
		},
	}
}
