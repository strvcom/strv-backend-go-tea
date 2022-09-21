package spec

import (
	"encoding/json"
	"reflect"

	"github.com/go-openapi/jsonpointer"
	specs "github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

type RequestBody struct {
	specs.Schema
	specs.Refable
	specs.VendorExtensible
	RequestBodyProps
}

type RequestBodyProps struct {
	Description string             `json:"description,omitempty"`
	Content     map[string]Content `json:"content,omitempty"`
	Required    bool               `json:"required"`
}

// JSONLookup implements an interface to customize json pointer lookup
func (rb RequestBody) JSONLookup(token string) (any, error) {
	if ex, ok := rb.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(rb.RequestBodyProps, token)
	return r, err
}

// UnmarshalJSON hydrates this items instance with the data from JSON
func (rb *RequestBody) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &rb.RequestBodyProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &rb.VendorExtensible); err != nil {
		return err
	}
	if reflect.DeepEqual(RequestBodyProps{}, rb.RequestBodyProps) {
		rb.RequestBodyProps = RequestBodyProps{}
	}
	return nil
}

// MarshalJSON converts this items object to JSON
func (rb RequestBody) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(rb.RequestBodyProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(rb.VendorExtensible)
	if err != nil {
		return nil, err
	}
	concated := swag.ConcatJSON(b1, b2)
	return concated, nil
}
