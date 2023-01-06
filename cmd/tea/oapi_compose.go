package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	cmderrors "go.strv.io/tea/pkg/errors"
	"go.strv.io/tea/pkg/openapi/load"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	yamlV2 "gopkg.in/yaml.v2"
)

var (
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
				e := &cmderrors.CommandError{}
				if errors.As(err, &e) {
					os.Exit(e.Code)
				}
				os.Exit(-1)
			}
		},
	}

	openapiComposeOptions = &OAPIComposeOptions{}
)

func init() {
	oapiCmd.AddCommand(composeCmd)

	composeCmd.Flags().StringVarP(&openapiComposeOptions.SourceFilePath, "source", "i", "", "path to OpenAPI schema to compose")
	composeCmd.Flags().StringVarP(&openapiComposeOptions.OutputFilePath, "output", "o", "", "path to OpenAPI output (defaults to STDOUT)")
	cobra.CheckErr(composeCmd.MarkFlagRequired("source"))
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
		return cmderrors.NewCommandError(err, cmderrors.CodeThirdParty)
	}

	exp, err := specDoc.Compose(&spec.ExpandOptions{
		RelativeBase: opts.SourceFilePath,
		SkipSchemas:  false,
	})
	if err != nil {
		return cmderrors.NewCommandError(err, cmderrors.CodeThirdParty)
	}

	b, err := json.Marshal(exp.Spec())
	if err == nil {
		d, err := swag.BytesToYAMLDoc(b)
		if err != nil {
			return cmderrors.NewCommandError(err, cmderrors.CodeThirdParty)
		}

		b, err = yamlV2.Marshal(d)
		if err != nil {
			return cmderrors.NewCommandError(err, cmderrors.CodeSerializing)
		}
	}

	if opts.OutputFilePath == "" {
		_, _ = fmt.Println(b)
	} else {
		err = os.WriteFile(opts.OutputFilePath, b, filePermissions)
		if err != nil {
			return cmderrors.NewCommandError(err, cmderrors.CodeIO)
		}
	}

	return nil
}
