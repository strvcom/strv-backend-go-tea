package main

import (
	"github.com/spf13/cobra"
)

var (
	genCmd = &cobra.Command{
		Use:   "gen",
		Short: "Code generating",
		Long: `This command provides a set of tools for code generating.

Example:
	tea gen -h

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}
)

func init() {
	rootCmd.AddCommand(genCmd)
}
