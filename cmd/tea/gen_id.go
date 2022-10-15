package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"

	cmderrors "go.strv.io/tea/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	// genIDCmd represents the id command
	genIDCmd = &cobra.Command{
		Use:   "id",
		Short: "Generate IDs",
		Long: `Generate useful methods for serialization/deserialization of IDs.

The generated methods are saved into an output file.

Example:
	tea gen id -i ./id.go -o ./id_gen.go

This command can also be used as an embedded go generator.

Example:
	//go:generate tea gen id -i ./id.go -o ./id_gen.go
	type (
		User         uint64
		RefreshToken uint64
	)
 `,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runGenerateIDs(genIDOptions.SourceFilePath, genIDOptions.OutputFilePath); err != nil {
				e := &cmderrors.CommandError{}
				if errors.As(err, &e) {
					os.Exit(e.Code)
				}
				os.Exit(-1)
			}
		},
	}

	genIDOptions = &GenIDOptions{}
)

func init() {
	genIDCmd.Flags().StringVarP(&genIDOptions.SourceFilePath, "source", "i", "", "path to file with id declarations")
	genIDCmd.Flags().StringVarP(&genIDOptions.OutputFilePath, "output", "o", "", "path to generated output")
	genCmd.AddCommand(genIDCmd)
}

type GenIDOptions struct {
	SourceFilePath string
	OutputFilePath string
}

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

const uuidTemplate = `
func unmarshalUUID(u *uuid.UUID, idTypeName string, data []byte) error {
	if err := u.UnmarshalText(data); err != nil {
		return fmt.Errorf("parse %q id value: %w", idTypeName, err)
	}
	return nil
}
{{ range .ids }}
func New{{ . }}() {{ . }} {
	return {{ . }}(uuid.New())
}

func (i {{ . }}) String() string {
	return uuid.UUID(i).String()
}

func (i {{ . }}) Empty() bool {
	return uuid.UUID(i) == uuid.Nil
}

func (i {{ . }}) MarshalText() ([]byte, error) {
	return []byte(uuid.UUID(i).String()), nil
}

func (i {{ . }}) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", uuid.UUID(i).String())), nil
}

func (i *{{ . }}) UnmarshalText(data []byte) error {
	return unmarshalUUID((*uuid.UUID)(i), "{{ . }}", data)
}

func (i *{{ . }}) UnmarshalJSON(data []byte) error {
	return unmarshalUUID((*uuid.UUID)(i), "{{ . }}", data)
}
{{ end }}`

// IDs stores multiple ID names under one kind.
type IDs map[string][]string

func (i IDs) generate() ([]byte, error) {
	output := &bytes.Buffer{}
	var genData []byte
	var err error

	for typ := range i {
		//nolint:gocritic,exhaustive
		switch typ {
		case "uint64":
			if genData, err = i.generateUint64ID(); err != nil {
				return nil, fmt.Errorf("generating uint64 ids: %w", err)
			}
		case "uuid.UUID":
			if genData, err = i.generateUUID(); err != nil {
				return nil, fmt.Errorf("generating uuid.UUID ids: %w", err)
			}
		}

		if _, err = output.Write(genData); err != nil {
			return nil, fmt.Errorf("writing ids: %w", err)
		}
	}

	return output.Bytes(), nil
}

func (i IDs) generateUint64ID() ([]byte, error) {
	ids, ok := i["uint64"]
	if !ok {
		return nil, nil
	}

	generatedOutput := &bytes.Buffer{}
	data := map[string][]string{
		"ids": ids,
	}

	t, err := template.New("uint64").Parse(uint64Template)
	if err != nil {
		return nil, err
	}
	if err = t.Execute(generatedOutput, data); err != nil {
		return nil, err
	}

	return generatedOutput.Bytes(), nil
}

func (i IDs) generateUUID() ([]byte, error) {
	ids, ok := i["uuid.UUID"]
	if !ok {
		return nil, nil
	}

	generatedOutput := &bytes.Buffer{}
	data := map[string][]string{
		"ids": ids,
	}

	t, err := template.New("uuid.UUID").Parse(uuidTemplate)
	if err != nil {
		return nil, err
	}
	if err = t.Execute(generatedOutput, data); err != nil {
		return nil, err
	}

	return generatedOutput.Bytes(), nil
}

func (i IDs) generateHeader() []byte {
	var d []byte
	d = append(d, "package id\n\n"...)
	d = append(d, "import (\n"...)
	d = append(d, "\t\"fmt\"\n\n"...)

	if _, ok := i["uuid.UUID"]; ok {
		d = append(d, "\t\"github.com/google/uuid\"\n"...)
	}

	return append(d, ")\n"...)
}

func supportedType(typ string) bool {
	if typ != "uint64" && typ != "uuid.UUID" {
		return false
	}
	return true
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
			varName := ts.Name.Name
			var typeName string
			// Primitive type
			typ, ok := ts.Type.(*ast.Ident)
			if ok {
				typeName = typ.String()
			}
			// Composite type
			compositeTyp, ok := ts.Type.(*ast.SelectorExpr)
			if ok {
				if leftSide, ok := compositeTyp.X.(*ast.Ident); ok {
					typeName = fmt.Sprintf("%s.%s", leftSide.Name, compositeTyp.Sel.Name)
				} else {
					typeName = compositeTyp.Sel.Name
				}
			}
			if ok = supportedType(typeName); !ok {
				if _, err = fmt.Printf("Unsupported id: name=%s, type=%s\n", varName, typeName); err != nil {
					return nil, err
				}
				continue
			}
			ids[typeName] = append(ids[typeName], varName)
		}
	}

	return ids, nil
}

func runGenerateIDs(sourceFilePath, outputFilePath string) error {
	if !strings.HasSuffix(sourceFilePath, ".go") {
		return cmderrors.NewCommandError(fmt.Errorf("invalid input file %q: expected .go file", sourceFilePath), cmderrors.CodeDependency)
	}
	if !strings.HasSuffix(outputFilePath, ".go") {
		return cmderrors.NewCommandError(fmt.Errorf("invalid output file %q: expected .go file", outputFilePath), cmderrors.CodeDependency)
	}

	ids, err := extractIDs(sourceFilePath)
	if err != nil {
		return cmderrors.NewCommandError(fmt.Errorf("extracting ids: %w", err), cmderrors.CodeCommand)
	}
	output, err := ids.generate()
	if err != nil {
		return cmderrors.NewCommandError(fmt.Errorf("generating ids: %w", err), cmderrors.CodeCommand)
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return cmderrors.NewCommandError(fmt.Errorf("creating output file: %w", err), cmderrors.CodeIO)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			_, _ = fmt.Printf("unable to close output file %q: %s", outputFilePath, err.Error())
		}
	}()

	if _, err = outputFile.Write(ids.generateHeader()); err != nil {
		return cmderrors.NewCommandError(fmt.Errorf("writing output header: %w", err), cmderrors.CodeIO)
	}
	if _, err = outputFile.Write(output); err != nil {
		return cmderrors.NewCommandError(fmt.Errorf("writing output data: %w", err), cmderrors.CodeIO)
	}

	return nil
}
