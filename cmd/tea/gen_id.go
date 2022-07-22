package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"go.strv.io/tea/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	// idCmd represents the id command
	idCmd = &cobra.Command{
		Use:   "id",
		Short: "Generate IDs",
		Long: `Generate useful methods for serialization/deserialization of IDs.

The generated methods are saved into an output file.

Example:
	tea gen id -i ./id.go -o ./id_gen.go

This command can also be used as an embedded go generator.

Example:
	//go:generate tea gen id -i ./id.go -o ./id_gen.go
 `,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runGenerateIDs(genIDOptions.SourceFilePath, genIDOptions.OutputFilePath); err != nil {
				os.Exit(err.(*errors.ErrCommand).Code)
			}
		},
	}

	genIDOptions = &GenIDOptions{}
)

func init() {
	idCmd.Flags().StringVarP(&genIDOptions.SourceFilePath, "input", "i", "", "path to file with id declarations")
	idCmd.Flags().StringVarP(&genIDOptions.OutputFilePath, "output", "o", "", "path to generated output")
	genCmd.AddCommand(idCmd)
}

type GenIDOptions struct {
	SourceFilePath string
	OutputFilePath string
}

const header = `package id

import (
	"fmt"
	"strconv"
)
`

const uint64Template = `
func unmarshalUint64(i *uint64, idTypeName string, data []byte) error {
	l := len(data)
	if l > 2 && data[0] == '"' && data[l-1] == '"' {
		data = data[1 : l-1]
	}
	uintNum, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("parse %q id value: %w", idTypeName, err)
	}
	*i = uintNum
	return nil
}
{{ range .ids }}
func (i {{ . }}) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", i)), nil
}

func (i {{ . }}) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%d\"", i)), nil
}

func (i *{{ . }}) UnmarshalText(data []byte) error {
	return unmarshalUint64((*uint64)(i), "{{ . }}", data)
}

func (i *{{ . }}) UnmarshalJSON(data []byte) error {
	return unmarshalUint64((*uint64)(i), "{{ . }}", data)
}
{{ end }}`

// IDs stores multiple ID names under one kind.
type IDs map[reflect.Kind][]string

func (i IDs) generate() ([]byte, error) {
	output := &bytes.Buffer{}
	for typ := range i {
		switch typ {
		case reflect.Uint64:
			genData, err := i.generateUint64ID()
			if _, err = output.Write(genData); err != nil {
				return nil, fmt.Errorf("writing uint64 ids: %w", err)
			}
		}
	}
	return output.Bytes(), nil
}

func (i IDs) generateUint64ID() ([]byte, error) {
	ids, ok := i[reflect.Uint64]
	if !ok {
		return nil, nil
	}

	generatedOutput := &bytes.Buffer{}
	data := map[string][]string{
		"ids": ids,
	}

	t, err := template.New(reflect.Uint64.String()).Parse(uint64Template)
	if err != nil {
		return nil, err
	}
	if err = t.Execute(generatedOutput, data); err != nil {
		return nil, err
	}

	return generatedOutput.Bytes(), nil
}

func supportedType(typ string) (reflect.Kind, bool) {
	switch typ {
	case reflect.Uint64.String():
		return reflect.Uint64, true
	default:
		return reflect.Invalid, false
	}
}

func extractIDs(filename string) (IDs, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	ids := IDs{}
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, s := range gd.Specs {
			ts, ok := s.(*ast.TypeSpec)
			if !ok || !ts.Name.IsExported() {
				continue
			}
			typ, ok := ts.Type.(*ast.Ident)
			if !ok {
				continue
			}
			t, ok := supportedType(typ.String())
			if !ok {
				fmt.Printf("Unsupported id: name=%s, type=%s\n", ts.Name.Name, typ.String())
				continue
			}
			ids[t] = append(ids[t], ts.Name.String())
		}
	}

	return ids, nil
}

func runGenerateIDs(sourceFilePath, outputFilePath string) error {
	inputFilePath, err := filepath.Abs(sourceFilePath)
	if err != nil {
		return errors.NewErrCommand(fmt.Errorf("absolute file path: %w", err), 2)
	}
	if !strings.HasSuffix(inputFilePath, ".go") {
		return errors.NewErrCommand(fmt.Errorf("invalid input file %s: expected .go file", inputFilePath), 2)
	}

	ids, err := extractIDs(inputFilePath)
	if err != nil {
		return errors.NewErrCommand(fmt.Errorf("extracting ids: %w", err), 3)
	}
	output, err := ids.generate()
	if err != nil {
		return errors.NewErrCommand(fmt.Errorf("generating ids: %w", err), 3)
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return errors.NewErrCommand(fmt.Errorf("creating output file: %w", err), 2)
	}

	if _, err = outputFile.Write([]byte(header)); err != nil {
		return errors.NewErrCommand(fmt.Errorf("writing output header: %w", err), 2)
	}
	if _, err = outputFile.Write(output); err != nil {
		return errors.NewErrCommand(fmt.Errorf("writing output data: %w", err), 2)
	}
	if err = outputFile.Close(); err != nil {
		return errors.NewErrCommand(fmt.Errorf("closing output file: %w", err), 2)
	}

	return nil
}
