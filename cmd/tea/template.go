package main

import (
	"bytes"
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
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// initCmd represents the init command
	repoTemplateCmd = &cobra.Command{
		Use:   "template",
		Short: "Template repository",
		Long: `Find and execute template files and folders in the repository.

By default, it executes all files ending with *.template in the local directory and its subdirectories.

Example:
	gokit repo init --recursive`,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(viper.Unmarshal(&repoTemplateCfg, decode.WithTagName("json")))

			if !rootOpt.SkipValidation {
				cobra.CheckErr(validate.Struct(&repoTemplateCfg))
			}
			cobra.CheckErr(runRepoTemplate(&repoTemplateCfg, &repoTemplateOpt))
		},
	}
	repoTemplateCfg RepoTemplateConfig  = RepoTemplateConfig{}
	repoTemplateOpt RepoTemplateOptions = RepoTemplateOptions{}
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
	repoTemplateCmd.Flags().StringArrayVar(&repoTemplateOpt.Vars,
		"set",
		[]string{}, "Set custom key-value pair passed to the template engine.",
	)
}

type RepoTemplateOptions struct {
	Dir       string
	Glob      string
	Vars      []string
	Recursive bool
	Remove    bool
	Suffix    string
}

type RepoTemplateConfig struct {
	RepoConfig `json:",squash"` //lint:ignore SA5008 squash is mapstructure directive

	Template struct {
		Vars []RepoTemplateVar `json:"vars" yaml:"vars"`
	} `json:"template" yaml:"template"`
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

	tVars := map[string]any{}
	if err := mapstructure.Decode(repoTemplateCfg.RootConfig, &tVars); err != nil {
		return fmt.Errorf("decode template vars: %w", err)
	}
	if err := mapstructure.Decode(repoTemplateCfg.RepoConfig, &tVars); err != nil {
		return fmt.Errorf("decode template vars: %w", err)
	}
	for _, v := range conf.Template.Vars {
		tVars[v.Name] = v.Data
	}
	// opts.Vars override conf.Vars
	for _, v := range opts.Vars {
		s := strings.Split(v, "=")
		if len(s) != 2 {
			return fmt.Errorf("invalid template var: %q", v)
		}
		tVars[s[0]] = s[1]
	}

	tFiles, err := matchFn(opts.Dir, opts.Glob)
	if err != nil {
		return fmt.Errorf("locate template files: %w", err)
	}
	for _, fPath := range tFiles {
		var b bytes.Buffer

		t, err := template.New(filepath.Base(fPath)).Funcs(sprig.GenericFuncMap()).ParseFiles(fPath)
		if err != nil {
			return fmt.Errorf("parse files: %w", err)
		}
		if err := t.Execute(&b, tVars); err != nil {
			return fmt.Errorf("execute template: %w", err)
		}

		p := strings.TrimSuffix(fPath, opts.Suffix)
		if err := ioutil.WriteFile(p, b.Bytes(), 0644); err != nil {
			return fmt.Errorf("write file %q: %w", p, err)
		}
	}

	if opts.Remove {
		if !rootOpt.Yes {
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

		for _, fPath := range tFiles {
			if err := os.Remove(fPath); err != nil {
				return fmt.Errorf("remove file %q: %w", fPath, err)
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
