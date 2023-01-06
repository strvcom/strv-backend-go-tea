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

const (
	filePermissions = 0600
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
		"skip-validatation", false, "whether to skip validation")
	rootCmd.PersistentFlags().BoolVar(&rootOpt.Yes,
		"yes",
		false, "confirm all prompts (defaults to false)",
	)
	rootCmd.PersistentFlags().BoolVar(&rootOpt.Verbose,
		"verbose", false, "verbose logging")

	validate = validator.New()
}

type RootOptions struct {
	ConfigPath     string
	SkipValidation bool
	Yes            bool
	Verbose        bool
}

type ContactInfo struct {
	Name  string `json:"name" yaml:"name" validate:"required"`
	Email string `json:"email" yaml:"email" validate:"required,email"`
	Phone string `json:"phone" yaml:"phone" validate:"omitempty,e164"`
}

// initRootConfig reads in config file and ENV variables if set.
func initRootConfig() {
	if len(os.Args) <= 1 {
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
	switch {
	case errors.As(err, &viper.ConfigFileNotFoundError{}):
		if rootOpt.Verbose {
			log(fmt.Sprintf("Config file not found: %s", viper.ConfigFileUsed()))
		}
	case err != nil:
		log(err.Error())
	default:
		if rootOpt.Verbose {
			log(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
		}
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
	switch {
	case errors.As(err, &viper.ConfigFileNotFoundError{}):
		if rootOpt.Verbose {
			log("Local config file not found. Skipping.")
		}
	case err != nil:
		log(err.Error())
	default:
		if rootOpt.Verbose {
			log(fmt.Sprintf("Using config file: %s", v.ConfigFileUsed()))
		}
		// Merge the local config into the existing default config. This will override
		// any default settings by the local changes.
		cobra.CheckErr(viper.MergeConfigMap(v.AllSettings()))
	}

	viper.AutomaticEnv() // read in environment variables that match
}

func log(msg string) {
	_, err := fmt.Fprintln(os.Stderr, msg)
	cobra.CheckErr(err)
}
