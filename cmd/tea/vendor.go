package main

import (
	"archive/tar"
	"bufio"
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

			packages, err := parseModFile()
			if err != nil {
				cmd.Println(fmt.Sprintf("Error parsing %s file: %v", GoModFile, err))
				os.Exit(1)
			}

			packagePath, err := getPathToPackages()
			if err != nil {
				cmd.Println(fmt.Sprintf("Error getting path to packages: %v", err))
				os.Exit(1)
			}

			//run go get <package>@<version>
			for _, p := range packages {
				command := exec.Command("go", "get", p.Name+"@"+p.Version)
				if err := command.Run(); err != nil {
					cmd.Println(fmt.Sprintf("Error running \"go get %s@%s\" command: %v", p.Name, p.Version, err))
					os.Exit(1)
				}

				if err := cp.Copy(fmt.Sprintf("%s/%s", packagePath, p.VersionedPackageName()),
					fmt.Sprintf("%s/%s", VendorFolder, p.VersionedPackageName()),
					cp.Options{AddPermission: ReadWritePermission},
				); err != nil {
					cmd.Println(fmt.Sprintf("Unable to copy packages to %s folder: %v", VendorFolder, err))
					os.Exit(1)
				}
			}

			//create tarball
			if err := CreateTar(VendorFolder); err != nil {
				cmd.Println(fmt.Sprintf("Unable to create tarball: %v", err))
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
	rootCmd.AddCommand(vendorCmd)
}

type VendorOptions struct {
	Filter string
	Output string
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

//TODO replace all `log.Fatal`s
func parseModFile() ([]InternalPackage, error) {
	f, err := os.Open(GoModFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s file: %w", GoModFile, err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	r, err := regexp.Compile(vendorOptions.Filter)
	if err != nil {
		return nil, fmt.Errorf("unable to compile regexp: %v", err)
	}

	var packages []InternalPackage
	for scanner.Scan() {
		line := scanner.Text()
		pName, ver, found := processLine(line, r)
		if !found {
			continue
		}
		packages = append(packages, InternalPackage{Name: pName, Version: ver})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to read %s file: %w", GoModFile, err)
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

var modReservedWords = []string{"require", "exclude", "replace"}

func processLine(line string, re *regexp.Regexp) (string, string, bool) {
	if strings.HasPrefix(line, "module") {
		return "", "", false
	}

	processedLine := line
	for _, word := range modReservedWords {
		processedLine = strings.TrimPrefix(processedLine, word+" ")
	}
	if re.MatchString(line) {
		processedLine = strings.TrimSpace(processedLine)
		if processedLine == "(" {
			return "", "", false
		}

		packageName, ver, found := strings.Cut(processedLine, " ")
		if !found {
			return "", "", false
		}

		return packageName, ver, true
	}

	return "", "", false
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
