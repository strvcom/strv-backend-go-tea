package main

import (
	"github.com/spf13/cobra"
)

var (
	oapiCmd = &cobra.Command{
		Use:     "openapi",
		Aliases: []string{"oapi"},
		Short:   "OpenAPI management tools",
		Long: `This command provides a set of tools to manage OpenAPI specifications.

Example:
	tea openapi -h

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}
)

func init() {
	rootCmd.AddCommand(oapiCmd)
}
