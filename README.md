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
	User             uint64
	RefreshToken     uuid.UUID
	DeviceIdentifier string
)
```
After triggering `go generate ./...` within an app, methods `MarshalText` and `UnmarshalText` along with other useful functions are generated.

[release]: https://img.shields.io/github/v/release/strvcom/strv-backend-go-tea
[codecov]: https://codecov.io/gh/strvcom/strv-backend-go-tea
[codecov-img]: https://codecov.io/gh/strvcom/strv-backend-go-tea/branch/master/graph/badge.svg?token=A7QFX32CFF
[license]: https://img.shields.io/github/license/strvcom/strv-backend-go-tea
