package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
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
		//TODO: Long
		Long: ` 
This command provides a set of tools to manage private Go packages.

Note that it is required to have a go in your $GOPATH and $GOROOT set up.

Example:
	tea vendor --filter=go.strv.io/*

Provided by ` + colorLinkSTRV,
		Run: func(cmd *cobra.Command, args []string) {
			if !isGoCommandAvailable() {
				cmd.Println("go command is not available")
				log.Fatal("go command is not available")
			}

			packages, err := parseModFile()
			if err != nil {
				cmd.Println(fmt.Sprintf("Error parsing %s file: %v", GoModFile, err))
				log.Fatal(err)
			}

			//run go get <package>@<version>
			for _, p := range packages {
				command := exec.Command("go", "get", p.Name+"@"+p.Version)
				if err := command.Run(); err != nil {
					cmd.Println(fmt.Sprintf("Error running \"go get %s@%s\" command: %v", p.Name, p.Version, err))
					log.Fatal(err)
				}
			}

			packagePath, err := getPathToPackages()
			if err != nil {
				cmd.Println(fmt.Sprintf("Error getting path to packages: %v", err))
				log.Fatal(err)
			}

			//copy folder to vendor folder
			for _, p := range packages {
				if err := cp.Copy(fmt.Sprintf("%s/%s@%s", packagePath, p.PackageName(), p.Version),
					fmt.Sprintf("%s/%s@%s", VendorFolder, p.PackageName(), p.Version),
					cp.Options{AddPermission: ReadWritePermission}); err != nil {
					cmd.Println(fmt.Sprintf("Unable to copy packages to %s folder: %v", VendorFolder, err))
					log.Fatal(err)
				}
			}

			//create tarball
			if err := CreateTar(VendorFolder); err != nil {
				log.Fatal(fmt.Sprintf("Unable to create tarball: %v", err))
			}
		},
	}

	vendorOptions         = &VendorOptions{}
	goPackageCharReplacer = strings.NewReplacer("A", "!a",
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

//PackageName returns the package name from the line
func (ip *InternalPackage) PackageName() string {
	return goPackageCharReplacer.Replace(ip.Name)
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
		fmt.Println(packageName, ver, found)
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
		log.Fatalln("Error writing archive:", err)
	}
	defer out.Close()

	gzw := gzip.NewWriter(out)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
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
