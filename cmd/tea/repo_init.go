package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.strv.io/tea/pkg/filecontent"

	"github.com/spf13/cobra"
)

const defaultDirectoryPermissions = 0755
const (
	pkgDir      = "pkg"
	cmdDir      = "cmd"
	modFileName = "go.mod"
)

var (
	repoInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize repository",
		Long: `Initialize the default Go repository structure.

Example:
	tea repo init`,
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(runRepoInit(&repoInitOpt))
		},
	}

	repoInitOpt = RepoInitOptions{}
)

func init() {
	repoCmd.AddCommand(repoInitCmd)

	repoInitCmd.Flags().StringVarP(&repoInitOpt.Dir,
		"dir",
		"d",
		".", "Directory to initialize (defaults to the current working directory)",
	)
	repoInitCmd.Flags().StringVarP(&repoInitOpt.Module,
		"module",
		"m",
		"strvcom/backend-go-example", "Name of the Go module to initialize",
	)
	repoInitCmd.Flags().StringVarP(&repoInitOpt.Author,
		"author",
		"a",
		"STRV", "Name of the author",
	)
	repoInitCmd.Flags().BoolVarP(&repoInitOpt.Replace,
		"replace",
		"r",
		true, "Whether to replace existing files.",
	)
}

type RepoInitOptions struct {
	Dir     string
	Module  string
	Author  string
	Replace bool
}

func runRepoInit(
	opts *RepoInitOptions,
) error {
	if err := initModule(opts.Dir, opts.Module); err != nil {
		return fmt.Errorf("initializing module: %w", err)
	}

	if err := initDefaultDirectories(opts.Dir); err != nil {
		return fmt.Errorf("initializing default directories: %w", err)
	}

	if err := initDefaultFiles(opts.Dir, opts.Replace); err != nil {
		return fmt.Errorf("initializing default files: %w", err)
	}

	if err := initTemplates(opts); err != nil {
		return fmt.Errorf("initializing templates: %w", err)
	}

	return nil
}

func initModule(dir, module string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current working directory: %w", err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			_, err = fmt.Fprintf(os.Stderr, "error: changing directory to %q: %v", cwd, err)
			if err != nil {
				panic(err)
			}
			os.Exit(1)
		}
	}()

	if err := os.MkdirAll(dir, defaultDirectoryPermissions); err != nil {
		return fmt.Errorf("creating directory %q: %w", dir, err)
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("changing directory to %q: %w", dir, err)
	}

	if _, err := os.Stat(modFileName); err == nil && repoInitOpt.Replace {
		if err := os.Remove(modFileName); err != nil {
			return fmt.Errorf("removing mod file: %w", err)
		}
	}
	cmd := exec.Command("go", "mod", "init", module)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func initTemplates(opts *RepoInitOptions) error {
	cmdCfg := &RepoTemplateConfig{
		Module:  opts.Module,
		Author:  opts.Author,
		Version: "0.1.0",
	}
	cmdOpt := &RepoTemplateOptions{
		Dir:       opts.Dir,
		Glob:      "*.template",
		Suffix:    ".template",
		Remove:    true,
		Recursive: true,
	}
	if err := runRepoTemplate(cmdCfg, cmdOpt); err != nil {
		return fmt.Errorf("running repo template: %w", err)
	}

	return nil
}

// initDefaultDirectories initializes the default Go repository structure.
func initDefaultDirectories(dir string) error {
	cmdDir := filepath.Join(dir, cmdDir)
	pkgDir := filepath.Join(dir, pkgDir)

	for _, d := range []string{
		cmdDir,
		pkgDir,
	} {
		if err := os.MkdirAll(d, defaultDirectoryPermissions); err != nil {
			return fmt.Errorf("creating directory %q: %w", d, err)
		}
	}

	return nil
}

func initDefaultFiles(dir string, replace bool) error {
	for _, f := range []struct {
		path    string
		content string
	}{
		{
			path:    filepath.Join(dir, ".cup.template"),
			content: filecontent.CupTemplate,
		},
		{
			path:    filepath.Join(dir, ".gitignore"),
			content: filecontent.Gitignore,
		},
		{
			path:    filepath.Join(dir, ".golangci.yml"),
			content: filecontent.Golangci,
		},
		{
			path:    filepath.Join(dir, "CHANGELOG"),
			content: filecontent.CHANGELOG,
		},
		{
			path:    filepath.Join(dir, "CODEOWNERS"),
			content: filecontent.CODEOWNERS,
		},
		{
			path:    filepath.Join(dir, "CONTRIBUTING"),
			content: filecontent.CONTRIBUTING,
		},
		{
			path:    filepath.Join(dir, "LICENSE"),
			content: filecontent.LICENSE,
		},
		{
			path:    filepath.Join(dir, "Makefile"),
			content: filecontent.Makefile,
		},
		{
			path:    filepath.Join(dir, "README.md"),
			content: filecontent.README,
		},
	} {
		if _, err := os.Stat(f.path); err == nil && !replace {
			continue
		}
		if err := os.WriteFile(f.path, []byte(f.content), defaultDirectoryPermissions); err != nil {
			return fmt.Errorf("creating file %q: %w", f.path, err)
		}
	}

	return nil
}
