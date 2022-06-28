package main

import "os"

func main() {
	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.
	if err := rootCmd.Execute(); err != nil {
		// TODO: Log the error.
		os.Exit(1)
	}
}
