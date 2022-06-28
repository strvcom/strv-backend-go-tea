package main

import (
	"errors"
	"fmt"
	"os"

	"go.strv.io/tea/pkg/termlink"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// rootCmd represents the base command when called without any subcommands.
	rootCmd = &cobra.Command{
		Use:   "tea",
		Short: "Go Tea!",
		Long: `Universal set of tools to make development in Go as simple as making a cup of tea.

Provided by ` + termlink.ColorLink("STRV", "https://strv.com", "red bold") + ".",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}
	cfgPath string

	validate     *validator.Validate //lint:ignore U1000 Ignore unused
	validateSkip bool
)

func init() {
	cobra.OnInitialize(
		initRootConfig,
	)

	rootCmd.PersistentFlags().StringVarP(&cfgPath,
		"config", "c", "", "config file (default is $HOME/.cup)")
	rootCmd.PersistentFlags().BoolVar(&validateSkip,
		"validate", false, "whether to skip validation")

	validate = validator.New()
}

type RootConfig struct {
	Module  string `json:"module" validate:"required"`
	Author  string `json:"author" validate:"required"`
	Version string `json:"version" validate:"required,semver"`

	Contributors []ContactInfo `json:"contributors,omitempty" validate:"omitempty"`
}

type ContactInfo struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"omitempty,e164"`
}

// initRootConfig reads in config file and ENV variables if set.
func initRootConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".cup" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".cup")

	// If a config file is found, read it in.
	err = viper.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// TODO: Add debug log here.
	} else if err != nil {
		// TODO: Add warning log here.
	} else {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	v := viper.New()
	if cfgPath != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgPath)
	} else {
		// Search local config in the root directory with name ".cup" (without extension).
		v.AddConfigPath("./")
		v.SetConfigType("yaml")
		v.SetConfigName(".cup")
	}

	// If a local config file is found, read it in and merge it with the default config.
	err = v.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// TODO: Add debug log here.
	} else if err != nil {
		// TODO: Add warning log here.
	} else {
		fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed())

		// Merge the local config into the existing default config. This will override
		// any default settings by the local changes.
		cobra.CheckErr(viper.MergeConfigMap(v.AllSettings()))
	}

	viper.AutomaticEnv() // read in environment variables that match
}
