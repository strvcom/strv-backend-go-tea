package spec

import (
	"encoding/json"

	"github.com/go-openapi/jsonpointer"
	specs "github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

type ContentProps struct {
	Description string                  `json:"description,omitempty"`
	Schema      *specs.Schema           `json:"schema,omitempty"`
	Headers     map[string]specs.Header `json:"headers,omitempty"`
	Examples    map[string]interface{}  `json:"examples,omitempty"`
}

type Content struct {
	specs.Refable
	ContentProps
	specs.VendorExtensible
}

// JSONLookup look up a value by the json property name
func (c Content) JSONLookup(token string) (interface{}, error) {
	if ex, ok := c.Extensions[token]; ok {
		return &ex, nil
	}
	if token == "$ref" {
		return &c.Ref, nil
	}
	ptr, _, err := jsonpointer.GetForToken(c.ContentProps, token)
	return ptr, err
}

// UnmarshalJSON hydrates this items instance with the data from JSON
func (c *Content) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &c.ContentProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &c.Refable); err != nil {
		return err
	}
	return json.Unmarshal(data, &c.VendorExtensible)
}

// MarshalJSON converts this items object to JSON
func (c Content) MarshalJSON() ([]byte, error) {
	var (
		b1  []byte
		err error
	)

	if c.Ref.String() == "" {
		// when there is no $ref, empty description is rendered as an empty string
		b1, err = json.Marshal(c.ContentProps)
	} else {
		// when there is $ref inside the schema, description should be omitempty-ied
		b1, err = json.Marshal(struct {
			Description string                  `json:"description,omitempty"`
			Schema      *specs.Schema           `json:"schema,omitempty"`
			Headers     map[string]specs.Header `json:"headers,omitempty"`
			Examples    map[string]interface{}  `json:"examples,omitempty"`
		}{
			Description: c.ContentProps.Description,
			Schema:      c.ContentProps.Schema,
			Examples:    c.ContentProps.Examples,
		})
	}
	if err != nil {
		return nil, err
	}

	b2, err := json.Marshal(c.Refable)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(c.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

func NewContent() *Content {
	return new(Content)
}
