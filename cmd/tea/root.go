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
	colorLinkSTRV = termlink.ColorLink("STRV", "https://strv.com", "red bold") + "."

	// rootCmd represents the base command when called without any subcommands.
	rootCmd = &cobra.Command{
		Use:   "tea",
		Short: "Go Tea!",
		Long: `Universal set of tools to make development in Go as simple as making a cup of tea.

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}
	rootOpt RootOptions

	validate *validator.Validate
)

func init() {
	cobra.OnInitialize(
		initRootConfig,
	)

	rootCmd.PersistentFlags().StringVarP(&rootOpt.ConfigPath,
		"config", "c", "", "config file (default is $HOME/.cup)")
	rootCmd.PersistentFlags().BoolVar(&rootOpt.SkipValidation,
		"validate", false, "whether to skip validation")
	rootCmd.PersistentFlags().BoolVar(&rootOpt.Yes,
		"yes",
		false, "Confirm all prompts (defaults to false)",
	)

	validate = validator.New()
}

type RootConfig struct {
	Module  string `json:"module" yaml:"module" validate:"required"`
	Author  string `json:"author" yaml:"author" validate:"required"`
	Version string `json:"version" yaml:"version" validate:"required,semver"`

	Contributors []ContactInfo `json:"contributors" yaml:"contributors,omitempty" validate:"omitempty"`
}

type RootOptions struct {
	ConfigPath     string
	SkipValidation bool
	Yes            bool
}

type ContactInfo struct {
	Name  string `json:"name" yaml:"name" validate:"required"`
	Email string `json:"email" yaml:"email" validate:"required,email"`
	Phone string `json:"phone" yaml:"phone" validate:"omitempty,e164"`
}

// initRootConfig reads in config file and ENV variables if set.
func initRootConfig() {
	if cobra.NoArgs(rootCmd, os.Args) != nil {
		return
	}

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
	v.SetConfigType("yaml")
	if rootOpt.ConfigPath != "" {
		// Use config file from the flag.
		v.SetConfigFile(rootOpt.ConfigPath)
	} else {
		// Search local config in the root directory with name ".cup" (without extension).
		v.AddConfigPath("./")
		v.SetConfigName(".cup")
	}

	// If a local config file is found, read it in and merge it with the default config.
	err = v.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		fmt.Fprintln(os.Stderr, "Local config file not found. Skipping.")
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed())

		// Merge the local config into the existing default config. This will override
		// any default settings by the local changes.
		cobra.CheckErr(viper.MergeConfigMap(v.AllSettings()))
	}

	viper.AutomaticEnv() // read in environment variables that match
}
