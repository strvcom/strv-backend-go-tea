package load

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	oapi "github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"go.strv.io/tea/pkg/openapi/spec"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

// Document represents a swagger spec document
type Document struct {
	spec         *spec.Swagger
	specFilePath string
	origSpec     *spec.Swagger
	//lint:ignore U1000 For external compatibility purposes.
	schema     *oapi.Schema
	raw        json.RawMessage
	pathLoader *loader
}

// JSONSpec load a spec from a json document
func JSONSpec(path string, options ...LoaderOption) (*Document, error) {
	data, err := JSONDoc(path)
	if err != nil {
		return nil, err
	}
	// convert to json
	return Analyzed(data, "", options...)
}

// Embedded returns a Document based on embedded specs. No analysis is required
func Embedded(orig, flat json.RawMessage, options ...LoaderOption) (*Document, error) {
	var origSpec, flatSpec spec.Swagger
	if err := json.Unmarshal(orig, &origSpec); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(flat, &flatSpec); err != nil {
		return nil, err
	}
	return &Document{
		raw:        orig,
		origSpec:   &origSpec,
		spec:       &flatSpec,
		pathLoader: loaderFromOptions(options),
	}, nil
}

// Spec load a new spec document from a local or remote path
func Spec(path string, options ...LoaderOption) (*Document, error) {

	ldr := loaderFromOptions(options)

	b, err := ldr.Load(path)
	if err != nil {
		return nil, err
	}

	document, err := Analyzed(b, "", options...)
	if err != nil {
		return nil, err
	}

	if document != nil {
		document.specFilePath = path
		document.pathLoader = ldr
	}

	return document, err
}

// Analyzed creates a new analyzed spec document for a root json.RawMessage.
func Analyzed(data json.RawMessage, version string, options ...LoaderOption) (*Document, error) {
	if version == "" {
		version = "2.0"
	}
	if version != "2.0" {
		return nil, fmt.Errorf("spec version %q is not supported", version)
	}

	raw, err := trimData(data) // trim blanks, then convert yaml docs into json
	if err != nil {
		return nil, err
	}

	swspec := new(spec.Swagger)
	if err = json.Unmarshal(raw, swspec); err != nil {
		return nil, err
	}

	origsqspec, err := cloneSpec(swspec)
	if err != nil {
		return nil, err
	}

	d := &Document{
		spec:       swspec,
		raw:        raw,
		origSpec:   origsqspec,
		pathLoader: loaderFromOptions(options),
	}

	return d, nil
}

func trimData(in json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(in)
	if len(trimmed) == 0 {
		return in, nil
	}

	if trimmed[0] == '{' || trimmed[0] == '[' {
		return trimmed, nil
	}

	// assume yaml doc: convert it to json
	yml, err := swag.BytesToYAMLDoc(trimmed)
	if err != nil {
		return nil, fmt.Errorf("analyzed: %v", err)
	}

	d, err := swag.YAMLToJSON(yml)
	if err != nil {
		return nil, fmt.Errorf("analyzed: %v", err)
	}

	return d, nil
}

// Compose composes the ref fields in the spec document and returns a new spec document
func (d *Document) Compose(options ...*oapi.ExpandOptions) (*Document, error) {
	swspec := new(spec.Swagger)
	if err := json.Unmarshal(d.raw, swspec); err != nil {
		return nil, err
	}

	var composeOptions *oapi.ExpandOptions
	if len(options) > 0 {
		composeOptions = options[0]
	} else {
		composeOptions = &oapi.ExpandOptions{
			RelativeBase: d.specFilePath,
		}
	}

	if composeOptions.PathLoader == nil {
		if d.pathLoader != nil {
			// use loader from Document options
			composeOptions.PathLoader = d.pathLoader.Load
		} else {
			// use package level loader
			composeOptions.PathLoader = loaders.Load
		}
	}

	if err := spec.ComposeSpec(swspec, composeOptions); err != nil {
		return nil, err
	}

	dd := &Document{
		spec:         swspec,
		specFilePath: d.specFilePath,
		raw:          d.raw,
		origSpec:     d.origSpec,
	}
	return dd, nil
}

// Spec returns the swagger spec object model
func (d *Document) Spec() *spec.Swagger {
	return d.spec
}

// Raw returns the raw swagger spec as json bytes
func (d *Document) Raw() json.RawMessage {
	return d.raw
}

// Pristine creates a new pristine document instance based on the input data
func (d *Document) Pristine() *Document {
	dd, _ := Analyzed(d.Raw(), d.spec.Swagger)
	dd.pathLoader = d.pathLoader
	return dd
}

func cloneSpec(src *spec.Swagger) (*spec.Swagger, error) {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(src); err != nil {
		return nil, err
	}

	var dst spec.Swagger
	if err := gob.NewDecoder(&b).Decode(&dst); err != nil {
		return nil, err
	}
	return &dst, nil
}
