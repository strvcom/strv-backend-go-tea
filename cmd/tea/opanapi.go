package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"go.strv.io/tea/pkg/errors"
	"go.strv.io/tea/pkg/openapi/load"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	oapiCmd = &cobra.Command{
		Use:     "openapi",
		Aliases: []string{"oapi"},
		Short:   "OpenAPI management tools",
		Long: `This command provides a set of tools to manage OpenAPI specifications.

Example:
	tea openapi -h`,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}

	// composeCmd represents the compose command
	composeCmd = &cobra.Command{
		Use:   "compose",
		Short: "Compose composite OpenAPI schema",
		Long: `Compose multiple OpenAPI specifications into a single schema.

The result is saved into an output file.

Example:
	tea openapi compose -i ./api/openapi_compose.yaml -o ./api/openapi.yaml

This command can also be used as an embedded go generator to embed OpenAPI specification.

Example:
	import _ "embed

	//go:generate tea openapi compose -i ./openapi_compose.yaml -o ./openapi.yaml
	//go:embed openapi.yaml
	var OpenAPI string
 `,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runOAPICompose(openapiComposeOptions); err != nil {
				os.Exit(err.(*errors.CommandError).Code)
			}
		},
	}

	openapiComposeOptions = &OAPIComposeOptions{}
)

func init() {
	composeCmd.Flags().StringVarP(&openapiComposeOptions.SourceFilePath, "source", "i", "", "path to OpenAPI schema to compose")
	composeCmd.Flags().StringVarP(&openapiComposeOptions.OutputFilePath, "output", "o", "", "path to OpenAPI output (defaults to STDOUT)")
	cobra.CheckErr(composeCmd.MarkFlagRequired("source"))

	oapiCmd.AddCommand(composeCmd)
	rootCmd.AddCommand(oapiCmd)
}

type OAPIComposeOptions struct {
	SourceFilePath string
	OutputFilePath string
}

func runOAPICompose(
	opts *OAPIComposeOptions,
) error {
	specDoc, err := load.Spec(opts.SourceFilePath)
	if err != nil {
		return errors.NewCommandError(err, 2)
	}

	exp, err := specDoc.Compose(&spec.ExpandOptions{
		RelativeBase: opts.SourceFilePath,
		SkipSchemas:  false,
	})
	if err != nil {
		return errors.NewCommandError(err, 3)
	}

	b, err := json.Marshal(exp.Spec())
	if err != nil {
		return errors.NewCommandError(err, 2)
	}
	d, err := swag.BytesToYAMLDoc(b)
	if err != nil {
		return errors.NewCommandError(err, 2)
	}
	b, err = yaml.Marshal(d)
	if err != nil {
		return errors.NewCommandError(err, 2)
	}

	if opts.OutputFilePath == "" {
		fmt.Println(string(b))
	} else {
		err = ioutil.WriteFile(opts.OutputFilePath, b, 0644)
		if err != nil {
			return errors.NewCommandError(err, 2)
		}
	}

	return nil
}
