package main

import (
	"errors"
	"os"

	cmderrors "go.strv.io/tea/pkg/errors"
)

func main() {
	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.
	if err := rootCmd.Execute(); err != nil {
		e := &cmderrors.CommandError{}
		if errors.As(err, &e) {
			os.Exit(e.Code)
		}
		os.Exit(1)
	}
}
