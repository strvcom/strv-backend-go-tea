package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
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
			if !isGoCommandAvailable() {
				cmd.Println("go command is not available")
				os.Exit(1)
			}

			re, err := regexp.Compile(vendorOptions.Filter)
			if err != nil {
				cmd.Println("unable to compile regexp: %v", err)
				os.Exit(1)
			}

			packages, err := getPackages(re)
			if err != nil {
				cmd.Println(fmt.Sprintf("error parsing %s file: %v", GoModFile, err))
				os.Exit(1)
			}

			if vendorOptions.DryRun {
				for _, p := range packages {
					cmd.Println(p.VersionedPackageName())
				}
				return
			}

			packagePath, err := getPathToPackages()
			if err != nil {
				cmd.Println(fmt.Sprintf("error getting path to packages: %v", err))
				os.Exit(1)
			}

			command := exec.Command("go", "mod", "download")
			if err := command.Run(); err != nil {
				cmd.Println(fmt.Sprintf("error running \"go mod download\" command: %v", err))
				os.Exit(1)
			}

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
			if err := CreateTar(VendorFolder); err != nil {
				cmd.Println(fmt.Sprintf("unable to create tarball: %v", err))
				os.Exit(1)
			}
		},
	}

	vendorOptions = &VendorOptions{}

	//go packages in go/pkg/mod folder has a syntax, where all uppercase letters are replaced with
	//exclamation mark and equivalent lowercase letter
	goPackageCharReplacer = strings.NewReplacer(
		"A", "!a",
		"B", "!b",
		"C", "!c",
		"D", "!d",
		"E", "!e",
		"F", "!f",
		"G", "!g",
		"H", "!h",
		"I", "!i",
		"J", "!j",
		"K", "!k",
		"L", "!l",
		"M", "!m",
		"N", "!n",
		"O", "!o",
		"P", "!p",
		"Q", "!q",
		"R", "!r",
		"S", "!s",
		"T", "!t",
		"U", "!u",
		"V", "!v",
		"W", "!w",
		"X", "!x",
		"Y", "!y",
		"Z", "!z",
	)
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

type InternalPackage struct {
	Name    string
	Version string
}

//VersionedPackageName returns the transformed package name as it is in modcache
func (ip *InternalPackage) VersionedPackageName() string {
	return fmt.Sprintf("%s@%s", goPackageCharReplacer.Replace(ip.Name), ip.Version)
}

func isGoCommandAvailable() bool {
	command := exec.Command("go", "version")
	//establish that we have "go" command
	if err := command.Run(); err != nil {
		return false
	}
	return true
}

func getPackages(re *regexp.Regexp) ([]InternalPackage, error) {
	command := exec.Command("go", "list", "-m", "all")
	out, err := command.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	packages := make([]InternalPackage, 0, len(lines))

	for _, line := range lines {
		splitLine := strings.Split(line, " ")
		if len(splitLine) != 2 || !re.MatchString(splitLine[0]) {
			continue
		}

		packages = append(packages, InternalPackage{
			Name:    splitLine[0],
			Version: splitLine[1],
		})
	}

	return packages, nil
}

func getPathToPackages() (string, error) {
	modcache := os.Getenv("GOMODCACHE")
	if modcache != "" {
		return modcache, nil
	}

	path := os.Getenv("GOPATH")
	if path != "" {
		return fmt.Sprintf("%s/pkg/mod", path), nil
	}

	_, err := os.Stat("~/go/pkg/mod")
	if err != nil {
		return "", err
	}

	return "~/go/pkg/mod", nil
}

// CreateTar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer
func CreateTar(src string) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf(fmt.Sprintf("Unable to tar files - %v", err))
	}

	outputPath := vendorOptions.Output
	if !strings.HasSuffix(outputPath, ".tar.gz") {
		outputPath = outputPath + ".tar.gz"
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error writing archive: %v", err)
	}
	defer out.Close()

	gzw := gzip.NewWriter(out)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.WalkDir(src, func(file string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		dInfo, err := d.Info()
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(dInfo, d.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}
