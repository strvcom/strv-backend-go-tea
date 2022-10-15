# STRV tea

![Latest release][release]
[![codecov][codecov-img]][codecov]
![GitHub][license]

Universal set of tools to make development in Go as simple as making a cup of tea.

## Installation
The most simple way to install tea is by Go native installation functionality. Example:
```shell
$ go install go.strv.io/tea/cmd/tea@${version:-latest}
```
Version points to a stable [release](https://github.com/strvcom/strv-backend-go-tea/releases).

## Commands
For detailed description of each command or global options, run the command with `-h` argument.

### version
Prints tea version and ends.

#### Example
```shell
$ tea version
$ 1.0.0
```

### gen
This command provides a set of tools for code generating.

Subcommand `id` generates useful methods for serialization/deserialization of IDs within Go apps. Example:
```shell
$ tea gen id -i ./id.go -o ./id_gen.go
```

This subcommand can also be used as an embedded go generator. Example:

id.go:
```go
package id

import (
	"github.com/google/uuid"
)

//go:generate tea gen id -i ./id.go -o ./id_gen.go

type (
	User         uint64
	RefreshToken uuid.UUID
)
```
After triggering `go generate ./...` within an app, methods `MarshalText`, `MarshalJSON`, `UnmarshalText` and `UnmarshalJSON` are generated.

### openapi
This command provides a set of tools to manage OpenAPI specifications.

Subcommand `compose` merges multiple OpenAPI specifications into a single schema. Example:
```shell
$ tea openapi compose -i ./api/openapi_compose.yaml -o ./api/openapi.yaml
```

This command can also be used as an embedded go generator to embed OpenAPI specification. Example:

```go
import _ "embed

//go:generate tea openapi compose -i ./openapi_compose.yaml -o ./openapi.yaml
//go:embed openapi.yaml
var OpenAPI string
```
After triggering `go generate ./...` within an app, `openapi.yaml` is generated. Afterthat, it is embedded into the app during build.

### repo
This command provides a set of tools to manage a local Go repository. Note that it is required to have a configured `.cup` file in the root of the repository
you wish to configure or have one in the home directory as a default.

Subcommand `template` finds and executes template files and folders in the repository. It is useful for creating a template repository
which will be later used for the creation of another project.
By default, it executes all files ending with `*.template` in the local directory and its subdirectories. Example:
```shell
$ tea repo template --recursive
```
.cup:
```yaml
repo:
  template:
    module: test
    author: Jane Doe
    values:
      - name: config
        data:
          port: 8080
    version: 0.1.0
```
test.template:
```yaml
module: {{ .Module }}
author: {{ .Author }}
config:
    port: "{{ .Values.config.port }}"
version: {{ .Version }}
```

[release]: https://img.shields.io/github/v/release/strvcom/strv-backend-go-tea
[codecov]: https://codecov.io/gh/strvcom/strv-backend-go-tea
[codecov-img]: https://codecov.io/gh/strvcom/strv-backend-go-tea/branch/master/graph/badge.svg?token=A7QFX32CFF
[license]: https://img.shields.io/github/license/strvcom/strv-backend-go-tea
