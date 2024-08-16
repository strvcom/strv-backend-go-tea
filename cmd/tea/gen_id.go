package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateIDs(genIDOptions.SourceFilePath, genIDOptions.OutputFilePath)
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

const stringTemplate = `{{ range .ids }}
func (i *{{ . }}) UnmarshalText(data []byte) error {
	*i = {{ . }}(data)
	return nil
}

func (i {{ . }}) MarshalText() ([]byte, error) {
	return []byte(i), nil
}
{{ end }}`

const uint64Template = `
func unmarshalUint64(i *uint64, idTypeName string, data []byte) error {
	uintNum, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("parsing %q id value: %w", idTypeName, err)
	}
	*i = uintNum
	return nil
}
{{ range .ids }}
func (i {{ . }}) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", i)), nil
}

func (i *{{ . }}) UnmarshalText(data []byte) error {
	return unmarshalUint64((*uint64)(i), "{{ . }}", data)
}
{{ end }}`

const uuidTemplate = `
func unmarshalUUID(u *uuid.UUID, idTypeName string, data []byte) error {
	if err := u.UnmarshalText(data); err != nil {
		return fmt.Errorf("parsing %q id value: %w", idTypeName, err)
	}
	return nil
}

func scanUUID(u *uuid.UUID, idTypeName string, data any) error {
	if err := u.Scan(data); err != nil {
		return fmt.Errorf("scanning %q id value: %w", idTypeName, err)
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

func (i *{{ . }}) UnmarshalText(data []byte) error {
	return unmarshalUUID((*uuid.UUID)(i), "{{ . }}", data)
}

func (i *{{ . }}) Scan(data any) error {
	return scanUUID((*uuid.UUID)(i), "{{ . }}", data)
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
		case "string":
			if genData, err = i.generateStringID(); err != nil {
				return nil, fmt.Errorf("generating string ids: %w", err)
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

func (i IDs) generateStringID() ([]byte, error) {
	ids, ok := i["string"]
	if !ok {
		return nil, nil
	}

	generatedOutput := &bytes.Buffer{}
	data := map[string][]string{
		"ids": ids,
	}

	t, err := template.New("string").Parse(stringTemplate)
	if err != nil {
		return nil, err
	}
	if err = t.Execute(generatedOutput, data); err != nil {
		return nil, err
	}

	return generatedOutput.Bytes(), nil
}

func (i IDs) generateHeader() []byte {
	// In case of a new type, add import dependencies to standardImports and externalImports.
	standardImports := make(map[string]struct{})
	externalImports := make(map[string]struct{})
	if _, ok := i["uint64"]; ok {
		standardImports["fmt"] = struct{}{}
		standardImports["strconv"] = struct{}{}
	}
	if _, ok := i["uuid.UUID"]; ok {
		standardImports["fmt"] = struct{}{}
		externalImports["github.com/google/uuid"] = struct{}{}
	}

	var (
		d                  []byte
		standardImportsLen = len(standardImports)
		externalImportsLen = len(externalImports)
	)

	d = append(d, "package id\n"...)
	if standardImportsLen == 0 && externalImportsLen == 0 {
		return d
	}

	d = append(d, "\nimport (\n"...)
	for _, v := range sortedMapKeys(standardImports) {
		d = append(d, fmt.Sprintf("\t\"%s\"\n", v)...)
	}

	if externalImportsLen == 0 {
		return append(d, ")\n"...)
	}

	d = append(d, "\n"...)
	for _, v := range sortedMapKeys(externalImports) {
		d = append(d, fmt.Sprintf("\t\"%s\"\n", v)...)
	}

	return append(d, ")\n"...)
}

func sortedMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func supportedType(typ string) bool {
	switch typ {
	case "uint64", "uuid.UUID", "string":
		return true
	}
	return false
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
