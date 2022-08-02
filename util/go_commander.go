package util

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type InternalPackage struct {
	Name    string
	Version string
}

var (
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

//VersionedPackageName returns the transformed package name as it is in modcache
func (ip *InternalPackage) VersionedPackageName() string {
	return fmt.Sprintf("%s@%s", goPackageCharReplacer.Replace(ip.Name), ip.Version)
}

type CommandExec interface {
	Output() ([]byte, error)
	Run() error
}

type GoCommand struct {
	CommandExec func(name string, args ...string) CommandExec
}

func (g *GoCommand) GetPackageList(re *regexp.Regexp) ([]InternalPackage, error) {
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

func (g *GoCommand) GetGoVersion() (string, error) {
	command := exec.Command("go", "version")
	out, err := command.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (g *GoCommand) ExecModDownload() error {
	command := exec.Command("go", "mod", "download")
	return command.Run()
}

func (g *GoCommand) GetMocCache() (string, error) {
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
