package spec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/go-openapi/jsonpointer"
	specs "github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

// Swagger this is the root document object for the API specification.
// It combines what previously was the Resource Listing and API Declaration (version 1.2 and earlier)
// together into one document.
//
// For more information: http://goo.gl/8us55a#swagger-object-
type Swagger struct {
	specs.VendorExtensible
	SwaggerProps
}

// JSONLookup look up a value by the json property name
func (s Swagger) JSONLookup(token string) (interface{}, error) {
	if ex, ok := s.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(s.SwaggerProps, token)
	return r, err
}

// MarshalJSON marshals this swagger structure to json
func (s Swagger) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.SwaggerProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

// UnmarshalJSON unmarshals a swagger spec from json
func (s *Swagger) UnmarshalJSON(data []byte) error {
	var sw Swagger
	if err := json.Unmarshal(data, &sw.SwaggerProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &sw.VendorExtensible); err != nil {
		return err
	}
	*s = sw
	return nil
}

// GobEncode provides a safe gob encoder for Swagger, including extensions
func (s Swagger) GobEncode() ([]byte, error) {
	var b bytes.Buffer
	raw := struct {
		Props SwaggerProps
		Ext   specs.VendorExtensible
	}{
		Props: s.SwaggerProps,
		Ext:   s.VendorExtensible,
	}
	err := gob.NewEncoder(&b).Encode(raw)
	return b.Bytes(), err
}

// GobDecode provides a safe gob decoder for Swagger, including extensions
func (s *Swagger) GobDecode(b []byte) error {
	var raw struct {
		Props SwaggerProps
		Ext   specs.VendorExtensible
	}
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(&raw)
	if err != nil {
		return err
	}
	s.SwaggerProps = raw.Props
	s.VendorExtensible = raw.Ext
	return nil
}

// SwaggerProps captures the top-level properties of an Api specification
//
// NOTE: validation rules
// - the scheme, when present must be from [http, https, ws, wss]
// - BasePath must start with a leading "/"
// - Paths is required
type SwaggerProps struct {
	OpenApi             string                       `json:"openapi,omitempty"`
	ID                  string                       `json:"id,omitempty"`
	Consumes            []string                     `json:"consumes,omitempty"`
	Produces            []string                     `json:"produces,omitempty"`
	Schemes             []string                     `json:"schemes,omitempty"`
	Swagger             string                       `json:"swagger,omitempty"`
	Info                *specs.Info                  `json:"info,omitempty"`
	Host                string                       `json:"host,omitempty"`
	BasePath            string                       `json:"basePath,omitempty"`
	Paths               *Paths                       `json:"paths"`
	Definitions         specs.Definitions            `json:"definitions,omitempty"`
	Parameters          map[string]specs.Parameter   `json:"parameters,omitempty"`
	Responses           map[string]Response          `json:"responses,omitempty"`
	SecurityDefinitions specs.SecurityDefinitions    `json:"securityDefinitions,omitempty"`
	Security            []map[string][]string        `json:"security,omitempty"`
	Tags                []specs.Tag                  `json:"tags,omitempty"`
	ExternalDocs        *specs.ExternalDocumentation `json:"externalDocs,omitempty"`
	Servers             []Server                     `json:"servers,omitempty"`
	Components          specs.SchemaOrArray          `json:"components,omitempty"`
}

type swaggerPropsAlias SwaggerProps

type gobSwaggerPropsAlias struct {
	Security []map[string]struct {
		List []string
		Pad  bool
	}
	Alias           *swaggerPropsAlias
	SecurityIsEmpty bool
}

// GobEncode provides a safe gob encoder for SwaggerProps, including empty security requirements
func (o SwaggerProps) GobEncode() ([]byte, error) {
	raw := gobSwaggerPropsAlias{
		Alias: (*swaggerPropsAlias)(&o),
	}

	var b bytes.Buffer
	if o.Security == nil {
		// nil security requirement
		err := gob.NewEncoder(&b).Encode(raw)
		return b.Bytes(), err
	}

	if len(o.Security) == 0 {
		// empty, but non-nil security requirement
		raw.SecurityIsEmpty = true
		raw.Alias.Security = nil
		err := gob.NewEncoder(&b).Encode(raw)
		return b.Bytes(), err
	}

	raw.Security = make([]map[string]struct {
		List []string
		Pad  bool
	}, 0, len(o.Security))
	for _, req := range o.Security {
		v := make(map[string]struct {
			List []string
			Pad  bool
		}, len(req))
		for k, val := range req {
			v[k] = struct {
				List []string
				Pad  bool
			}{
				List: val,
			}
		}
		raw.Security = append(raw.Security, v)
	}

	err := gob.NewEncoder(&b).Encode(raw)
	return b.Bytes(), err
}

// GobDecode provides a safe gob decoder for SwaggerProps, including empty security requirements
func (o *SwaggerProps) GobDecode(b []byte) error {
	var raw gobSwaggerPropsAlias

	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(&raw)
	if err != nil {
		return err
	}
	if raw.Alias == nil {
		return nil
	}

	switch {
	case raw.SecurityIsEmpty:
		// empty, but non-nil security requirement
		raw.Alias.Security = []map[string][]string{}
	case len(raw.Alias.Security) == 0:
		// nil security requirement
		raw.Alias.Security = nil
	default:
		raw.Alias.Security = make([]map[string][]string, 0, len(raw.Security))
		for _, req := range raw.Security {
			v := make(map[string][]string, len(req))
			for k, val := range req {
				v[k] = make([]string, 0, len(val.List))
				v[k] = append(v[k], val.List...)
			}
			raw.Alias.Security = append(raw.Alias.Security, v)
		}
	}

	*o = *(*SwaggerProps)(raw.Alias)
	return nil
}

// vim:set ft=go noet sts=2 sw=2 ts=2:
