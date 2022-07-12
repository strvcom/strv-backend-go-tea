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

	"github.com/Masterminds/sprig/v3"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

var (
	repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Repository management tools",
		Long: `This command provides a set of tools to manage local Go repository.

Note that it is required to have a configured .cup file in the root of the repository
you wish to configure.

Example:
	tea repo -h`,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(cmd.Usage())
		},
	}

	// initCmd represents the init command
	repoTemplateCmd = &cobra.Command{
		Use:   "template",
		Short: "Template repository",
		Long: `Find and execute template files and folders in the repository.

By default, it executes all files ending with *.template in the local directory and its subdirectories.

Example:
	tea repo template --recursive`,
		Run: func(cmd *cobra.Command, args []string) {
			if !rootOpt.SkipValidation {
				cobra.CheckErr(validate.Struct(&repoTemplateCfg))
			}
			cobra.CheckErr(runRepoTemplate(&repoTemplateCfg, &repoTemplateOpt))
		},
	}

	repoTemplateCfg = RepoTemplateConfig{}
	repoTemplateOpt = RepoTemplateOptions{}
)

func init() {
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
		[]string{}, "Set custom key-value pair passed to the template engine",
	)

	repoCmd.AddCommand(repoTemplateCmd)
	rootCmd.AddCommand(repoCmd)
}

type RepoConfig struct {
	RepoTemplateConfig RepoTemplateConfig `json:"template" yaml:"template"`
}

type RepoTemplateConfig struct {
	Module       string         `json:"module" yaml:"module" mapstructure:"Module" validate:"required"`
	Author       string         `json:"author" yaml:"author" mapstructure:"Author" validate:"required"`
	Version      string         `json:"version" yaml:"version" mapstructure:"Version" validate:"required,semver"`
	Contributors []ContactInfo  `json:"contributors" yaml:"contributors" mapstructure:"Contributors"`
	Vars         map[string]any `json:"vars" yaml:"vars" mapstructure:"Vars"`
}

type RepoTemplateOptions struct {
	Dir       string
	Glob      string
	Vars      []string
	Recursive bool
	Remove    bool
	Suffix    string
}

func runRepoTemplate(conf *RepoTemplateConfig, opts *RepoTemplateOptions) error {
	tData, err := templateData(conf, opts)
	if err != nil {
		return fmt.Errorf("generating template data: %w", err)
	}

	var matchFn func(string, string) ([]string, error)
	if opts.Recursive {
		matchFn = func(dir, glob string) ([]string, error) { return recursiveGlob(dir, glob) }
	} else {
		matchFn = func(dir, glob string) ([]string, error) { return filepath.Glob(filepath.Join(dir, glob)) }
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
		if err := t.Execute(&b, tData); err != nil {
			return fmt.Errorf("execute template: %w", err)
		}

		p := strings.TrimSuffix(fPath, opts.Suffix)
		if err := ioutil.WriteFile(p, b.Bytes(), 0644); err != nil {
			return fmt.Errorf("write file %q: %w", p, err)
		}
	}

	if opts.Remove {
		if err := removeFiles(tFiles); err != nil {
			return fmt.Errorf("remove file: %w", err)
		}
	}

	return nil
}

func templateData(conf *RepoTemplateConfig, opts *RepoTemplateOptions) (map[string]any, error) {
	tData := map[string]any{}
	if err := mapstructure.Decode(conf, &tData); err != nil {
		return nil, fmt.Errorf("decode template vars: %w", err)
	}

	// opts.Vars override conf.Vars
	if tData["Vars"].(map[string]any) == nil {
		tData["Vars"] = make(map[string]any)
	}
	vars := tData["Vars"].(map[string]any)
	for _, v := range opts.Vars {
		s := strings.Split(v, "=")
		if len(s) != 2 {
			return nil, fmt.Errorf("invalid template var: %q", v)
		}
		overrideVarByOpts(vars, s[0], s[1])
	}

	return tData, nil
}

func overrideVarByOpts(vars map[string]any, key, value string) {
	index := strings.Index(key, ".")
	if index == -1 {
		vars[key] = value
		return
	}

	k := key[:index]
	if _, ok := vars[k]; !ok {
		vars[k] = map[string]any{}
	}
	overrideVarByOpts(vars[k].(map[string]any), key[index+1:], value)
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

func removeFiles(files []string) error {
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
	for _, fPath := range files {
		if err := os.Remove(fPath); err != nil {
			return err
		}
	}
	return nil
}
