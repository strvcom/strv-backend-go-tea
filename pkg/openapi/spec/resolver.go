package spec

import (
	"fmt"

	specs "github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

func resolveAnyWithBase(root interface{}, ref *specs.Ref, result interface{}, options *specs.ExpandOptions) error {
	options = optionsOrDefault(options)
	resolver := defaultSchemaLoader(root, options, nil, nil)

	if err := resolver.Resolve(ref, result, options.RelativeBase); err != nil {
		return err
	}

	return nil
}

// ResolveRefWithBase resolves a reference against a context root with preservation of base path
func ResolveRefWithBase(root interface{}, ref *specs.Ref, options *specs.ExpandOptions) (*specs.Schema, error) {
	result := new(specs.Schema)

	if err := resolveAnyWithBase(root, ref, result, options); err != nil {
		return nil, err
	}

	return result, nil
}

// ResolveRef resolves a reference for a schema against a context root
// ref is guaranteed to be in root (no need to go to external files)
//
// ResolveRef is ONLY called from the code generation module
func ResolveRef(root interface{}, ref *specs.Ref) (*specs.Schema, error) {
	res, _, err := ref.GetPointer().Get(root)
	if err != nil {
		return nil, err
	}

	switch sch := res.(type) {
	case specs.Schema:
		return &sch, nil
	case *specs.Schema:
		return sch, nil
	case map[string]interface{}:
		newSch := new(specs.Schema)
		if err = swag.DynamicJSONToStruct(sch, newSch); err != nil {
			return nil, err
		}
		return newSch, nil
	default:
		return nil, fmt.Errorf("type: %T: %w", sch, ErrUnknownTypeForReference)
	}
}

// ResolveParameterWithBase resolves a parameter reference against a context root and base path
func ResolveParameterWithBase(root interface{}, ref specs.Ref, options *specs.ExpandOptions) (*specs.Parameter, error) {
	result := new(specs.Parameter)

	if err := resolveAnyWithBase(root, &ref, result, options); err != nil {
		return nil, err
	}

	return result, nil
}

// ResolveParameter resolves a parameter reference against a context root
func ResolveParameter(root interface{}, ref specs.Ref) (*specs.Parameter, error) {
	return ResolveParameterWithBase(root, ref, nil)
}

// ResolveResponseWithBase resolves response a reference against a context root and base path
func ResolveResponseWithBase(root interface{}, ref specs.Ref, options *specs.ExpandOptions) (*Response, error) {
	result := new(Response)

	err := resolveAnyWithBase(root, &ref, result, options)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ResolveResponse resolves response a reference against a context root
func ResolveResponse(root interface{}, ref specs.Ref) (*Response, error) {
	return ResolveResponseWithBase(root, ref, nil)
}

// ResolvePathItemWithBase resolves response a path item against a context root and base path
func ResolvePathItemWithBase(root interface{}, ref specs.Ref, options *specs.ExpandOptions) (*PathItem, error) {
	result := new(PathItem)

	if err := resolveAnyWithBase(root, &ref, result, options); err != nil {
		return nil, err
	}

	return result, nil
}

// ResolvePathItem resolves response a path item against a context root and base path
//
// Deprecated: use ResolvePathItemWithBase instead
func ResolvePathItem(root interface{}, ref specs.Ref, options *specs.ExpandOptions) (*PathItem, error) {
	return ResolvePathItemWithBase(root, ref, options)
}

// ResolveItemsWithBase resolves parameter items reference against a context root and base path.
//
// NOTE: stricly speaking, this construct is not supported by Swagger 2.0.
// Similarly, $ref are forbidden in response headers.
func ResolveItemsWithBase(root interface{}, ref specs.Ref, options *specs.ExpandOptions) (*specs.Items, error) {
	result := new(specs.Items)

	if err := resolveAnyWithBase(root, &ref, result, options); err != nil {
		return nil, err
	}

	return result, nil
}

// ResolveItems resolves parameter items reference against a context root and base path.
//
// Deprecated: use ResolveItemsWithBase instead
func ResolveItems(root interface{}, ref specs.Ref, options *specs.ExpandOptions) (*specs.Items, error) {
	return ResolveItemsWithBase(root, ref, options)
}
