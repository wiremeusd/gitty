package main

import (
	"github.com/spf13/cobra"

	"github.com/wiremeusd/gitty/internal/auth"
	"github.com/wiremeusd/gitty/internal/credential"
)

func newCredentialCmd() *cobra.Command {
	var account string
	cmd := &cobra.Command{
		Use:    "credential <get|store|erase>",
		Short:  "Internal command: git credential helper",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return credential.Run(account, args[0], cmd.InOrStdin(), cmd.OutOrStdout(), auth.Token)
		},
	}
	cmd.Flags().StringVar(&account, "account", "", "gitty account name")
	_ = cmd.MarkFlagRequired("account")
	return cmd
}
