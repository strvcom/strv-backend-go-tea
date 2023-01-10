package filecontent

import _ "embed"

var (
	//go:embed _assets/.cup.template
	CupTemplate string
	//go:embed _assets/.gitignore
	Gitignore string
	//go:embed _assets/.golangci.yml
	Golangci string
	//go:embed _assets/CHANGELOG.md
	CHANGELOG string
	//go:embed _assets/CODEOWNERS
	CODEOWNERS string
	//go:embed _assets/CONTRIBUTING.md
	CONTRIBUTING string
	//go:embed _assets/LICENSE
	LICENSE string
	//go:embed _assets/Makefile
	Makefile string
	//go:embed _assets/README.md
	README string
)
