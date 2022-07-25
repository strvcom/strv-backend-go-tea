package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.strv.io/tea/pkg/decode"

	"github.com/Masterminds/sprig/v3"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	repoTemplateCmd = &cobra.Command{
		Use:   "template",
		Short: "Template repository",
		Long: `Find and execute template files and folders in the repository.

By default, it executes all files ending with *.template in the local directory and its subdirectories.

Example:
	tea repo template --recursive`,
		Run: func(cmd *cobra.Command, args []string) {
			v := viper.Sub(repoTemplateCfg.Key())
			if v != nil {
				cobra.CheckErr(v.Unmarshal(
					&repoTemplateCfg,
					decode.WithTagName("json"),
					decode.WithDecodeHook(decode.UnmarshalJSONHookFunc),
				))
			}

			if !rootOpt.SkipValidation {
				cobra.CheckErr(validate.Struct(repoTemplateCfg))
			}
			cobra.CheckErr(runRepoTemplate(&repoTemplateCfg, &repoTemplateOpt))
		},
	}

	repoTemplateCfg = RepoTemplateConfig{}
	repoTemplateOpt = RepoTemplateOptions{}
)

func init() {
	repoCmd.AddCommand(repoTemplateCmd)

	repoTemplateCmd.Flags().StringVarP(&repoTemplateOpt.Dir,
		"dir",
		"d",
		".", "Directory with template files (defaults to the current working directory)",
	)
	repoTemplateCmd.Flags().StringVarP(&repoTemplateOpt.Glob,
		"glob",
		"g",
		"*.template", "Glob pattern for template files (defaults to *.template)",
	)
	repoTemplateCmd.Flags().BoolVarP(&repoTemplateOpt.Recursive,
		"recursive",
		"R",
		false, "Whether to search templates recursively in subdirectories (defaults to false)",
	)
	repoTemplateCmd.Flags().BoolVar(&repoTemplateOpt.Remove,
		"remove",
		false, "Whether to remove templates after templating (defaults to false)",
	)
	repoTemplateCmd.Flags().StringVarP(&repoTemplateOpt.Suffix,
		"suffix",
		"s",
		".template", "Template file suffix. This suffix will be trimmed (defaults to .template)",
	)
	repoTemplateCmd.Flags().StringArrayVar(&repoTemplateOpt.Values,
		"set",
		[]string{}, "Set custom key-value pair passed to the template engine.",
	)
}

type RepoTemplateOptions struct {
	Dir       string
	Glob      string
	Remove    bool
	Suffix    string
	Values    []string
	Recursive bool
}

type RepoTemplateConfig struct {
	Module       string                   `json:"module" yaml:"module" validate:"required"`
	Author       string                   `json:"author" yaml:"author" validate:"required"`
	Version      string                   `json:"version" yaml:"version" validate:"required,semver"`
	Values       RepoTemplateValuesMapper `json:"values" yaml:"vars"`
	Contributors []ContactInfo            `json:"contributors" yaml:"contributors"`
}

func (*RepoTemplateConfig) Key() string {
	return "repo.template"
}

type RepoTemplateValuesMapper struct {
	orig []RepoTemplateVar
	data map[string]any
}

func (r *RepoTemplateValuesMapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.data)
}

func (r *RepoTemplateValuesMapper) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.orig); err != nil {
		return err
	}
	r.data = map[string]any{}
	for _, v := range r.orig {
		r.data[v.Name] = v.Data
	}
	return nil
}

type RepoTemplateVar struct {
	Name string `json:"name" yaml:"name"`
	Data any    `json:"data" yaml:"data"`
}

func runRepoTemplate(
	conf *RepoTemplateConfig,
	opts *RepoTemplateOptions,
) error {
	var matchFn func(string, string) ([]string, error)

	if opts.Recursive {
		matchFn = func(dir, glob string) ([]string, error) { return recursiveGlob(dir, glob) }
	} else {
		matchFn = func(dir, glob string) ([]string, error) { return filepath.Glob(filepath.Join(dir, glob)) }
	}

	tData, err := toTemplateData(conf, opts)
	if err != nil {
		return fmt.Errorf("capitalizing keys: %w", err)
	}
	if err := toCapitalKeys(tData, 0, 0); err != nil {
		return fmt.Errorf("capitalizing keys: %w", err)
	}

	tFiles, err := matchFn(opts.Dir, opts.Glob)
	if err != nil {
		return fmt.Errorf("locating template files: %w", err)
	}
	for _, fPath := range tFiles {
		var b bytes.Buffer

		t, err := template.New(filepath.Base(fPath)).Funcs(sprig.GenericFuncMap()).ParseFiles(fPath)
		if err != nil {
			return fmt.Errorf("parsing file: %w", err)
		}
		if err := t.Execute(&b, tData); err != nil {
			return fmt.Errorf("execute template: %w", err)
		}

		p := strings.TrimSuffix(fPath, opts.Suffix)
		if err := ioutil.WriteFile(p, b.Bytes(), 0644); err != nil {
			return fmt.Errorf("writing file %q: %w", p, err)
		}
	}

	if opts.Remove {
		if err := removeTemplateFiles(tFiles, !rootOpt.Yes); err != nil {
			return fmt.Errorf("removing template files: %w", err)
		}
	}

	return nil
}

func toTemplateData(
	conf *RepoTemplateConfig,
	opts *RepoTemplateOptions,
) (map[string]any, error) {
	tData, err := decode.ToMap(conf)
	if err != nil {
		return nil, fmt.Errorf("decoding template data: %w", err)
	}
	if _, ok := tData["values"]; !ok {
		return nil, errors.New("missing key 'values'")
	}
	vals, ok := tData["values"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid type of key 'values': expected 'map[string]any', got '%t'", tData["values"])
	}
	if err := overrideValuesByOpts(vals, opts.Values); err != nil {
		return nil, fmt.Errorf("overriding template data: %w", err)
	}

	return tData, nil
}

func removeTemplateFiles(files []string, skipPrompt bool) error {
	if skipPrompt {
		t := &promptui.PromptTemplates{
			Invalid: `{{ "Operation aborted" | red }}`,
			Success: "",
			Confirm: `{{ . | bold }} {{ "[y/N]" | faint }}`,
		}

		prompt := promptui.Prompt{
			Label:     "All template files have been templated. Do you want to remove them?",
			Templates: t,
			IsConfirm: true,
		}
		_, err := prompt.Run()
		if err != nil {
			// Immediately exit if the user doesn't want to remove the files.
			os.Exit(1)
		}
	}

	for _, fPath := range files {
		if err := os.Remove(fPath); err != nil {
			return fmt.Errorf("removing file %q: %w", fPath, err)
		}
	}

	return nil
}

func overrideValuesByOpts(data map[string]any, valueOpts []string) error {
	v := viper.New()
	for _, to := range valueOpts {
		s := strings.Split(to, "=")
		if len(s) != 2 {
			return fmt.Errorf("invalid template value: %q", to)
		}

		key, value := s[0], s[1]
		if key[0] == '.' {
			return fmt.Errorf("invalid template key: %q", key)
		}
		v.Set(key, value)
	}
	for k, v := range v.AllSettings() {
		data[k] = v
	}

	return nil
}

func toCapitalKeys(data map[string]any, depth, maxDepth int) error {
	for k, v := range data {
		if depth > maxDepth {
			return nil
		}

		newKey := strings.ToUpper(string(k[0])) + k[1:]
		if newKey != k {
			data[newKey] = v
			delete(data, k)
		}

		if v, ok := v.(map[string]any); ok {
			if err := toCapitalKeys(v, depth+1, maxDepth); err != nil {
				return err
			}
		}
	}
	return nil
}

func recursiveGlob(dir, glob string) (matches []string, err error) {
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, _ error) error {
		if !d.IsDir() {
			return nil
		}
		m, err := filepath.Glob(filepath.Join(path, glob))
		if err != nil {
			return err
		}
		matches = append(matches, m...)
		return nil
	})
	return matches, err
}
