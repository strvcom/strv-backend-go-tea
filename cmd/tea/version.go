package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version and exit",
		Long: `This command displays current version and exits.

Example:
	tea version

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			version := "unknown"
			buildInfo, ok := debug.ReadBuildInfo()
			if ok {
				version = buildInfo.Main.Version
			}
			_, _ = fmt.Println(version)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
