package main

import (
	"github.com/spf13/cobra"
)

var (
	repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Repository management tools",
		Long: `This command provides a set of tools to manage local Go repository.

Note that it is required to have a configured .cup file in the root of the repository
you wish to configure.

Example:
	tea repo -h

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}
)

func init() {
	rootCmd.AddCommand(repoCmd)
}
