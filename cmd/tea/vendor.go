package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"go.strv.io/tea/util"
)

const (
	ReadWritePermission = 0666
	VendorFolder        = ".vendor"
	GoModFile           = "go.mod"
)

var (
	vendorCmd = &cobra.Command{
		Use:   "vendor",
		Short: "Vendor private packages",
		Long: ` 
This command provides a set of tools to manage private Go packages.

Note that it is required to have a go in your $GOPATH and $GOROOT set up.

Example:
	tea vendor --filter=go.strv.io

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			goCmd := &util.GoCommand{
				CommandExec: exec.Command,
			}
			if !isGoCommandAvailable(goCmd) {
				cmd.Println("go command is not available")
				os.Exit(1)
			}

			packages, err := getFilteredPackages(vendorOptions.Filter, goCmd)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			if vendorOptions.DryRun {
				for _, p := range packages {
					cmd.Println(p.VersionedPackageName())
				}
				return
			}

			packagePath, err := goCmd.GetMocCache()
			if err != nil {
				cmd.Println(fmt.Sprintf("error getting path to packages: %v", err))
				os.Exit(1)
			}

			cmd.Println("Downloading packages via \"go mod download\"...")
			if err := goCmd.ExecModDownload(); err != nil {
				cmd.Println(fmt.Sprintf("error running \"go mod download\" command: %v", err))
				os.Exit(1)
			}

			cmd.Println("Copying packages to .vendor folder...")
			for _, p := range packages {
				if err := cp.Copy(fmt.Sprintf("%s/%s", packagePath, p.VersionedPackageName()),
					fmt.Sprintf("%s/%s", VendorFolder, p.VersionedPackageName()),
					cp.Options{AddPermission: ReadWritePermission},
				); err != nil {
					cmd.Println(fmt.Sprintf("unable to copy packages to %s folder: %v", VendorFolder, err))
					os.Exit(1)
				}
			}

			//create tarball
			cmd.Println("Creating tarball with private packages...")
			if err := util.CreateTar(VendorFolder, vendorOptions.Output); err != nil {
				cmd.Println(fmt.Sprintf("unable to create tarball: %v", err))
				os.Exit(1)
			}

			//remove .vendor folder
			cmd.Println("Removing .vendor folder...")
			if err := os.RemoveAll(VendorFolder); err != nil {
				cmd.Println(fmt.Sprintf("unable to remove .vendor folder: %v", err))
				os.Exit(1)
			}
		},
	}

	vendorOptions = &VendorOptions{}
)

func init() {
	vendorCmd.Flags().StringVarP(&vendorOptions.Filter, "filter", "f", "", "Filter the packages to vendor")
	vendorCmd.Flags().StringVarP(&vendorOptions.Output, "output", "o", "vendor", "Output file")
	vendorCmd.Flags().BoolVar(&vendorOptions.DryRun, "dry-run", false, "Dry run")
	rootCmd.AddCommand(vendorCmd)
}

type VendorOptions struct {
	Filter string
	Output string
	DryRun bool
}

func isGoCommandAvailable(goCmd *util.GoCommand) bool {
	if _, err := goCmd.GetGoVersion(); err != nil {
		return false
	}

	return true
}

func getFilteredPackages(filter string, goCmd *util.GoCommand) ([]util.InternalPackage, error) {
	re, err := regexp.Compile(filter)
	if err != nil {
		return nil, fmt.Errorf("unable to compile regexp: %v", err)
	}

	packages, err := goCmd.GetPackageList(re)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s file: %v", GoModFile, err)
	}

	return packages, nil
}
