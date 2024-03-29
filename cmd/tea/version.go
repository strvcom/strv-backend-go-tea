package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.0.0" // version is set during build

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version and exit",
		Long: `This command displays current version and exits.

Example:
	tea version

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Println(version)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
