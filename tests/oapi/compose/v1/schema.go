package v1

import (
	_ "embed"
)

//go:generate tea openapi compose -i openapi_compose.yaml -o openapi.yaml
//go:embed openapi.yaml
var openAPI string
