package main

import (
	"github.com/spf13/cobra"

	"go.strv.io/tea/pkg/termlink"
)

var (
	colorLinkSTRV = termlink.ColorLink("STRV", "https://strv.com", "red bold") + "."

	// rootCmd represents the base command when called without any subcommands.
	rootCmd = &cobra.Command{
		Use:   "tea",
		Short: "Go Tea!",
		Long: `Universal set of tools to make development in Go as simple as making a cup of tea.

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, _ []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}
)
