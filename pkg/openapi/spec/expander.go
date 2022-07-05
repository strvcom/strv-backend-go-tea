package spec

import (
	"fmt"

	specs "github.com/go-openapi/spec"
)

func optionsOrDefault(opts *specs.ExpandOptions) *specs.ExpandOptions {
	if opts != nil {
		clone := *opts // shallow clone to avoid internal changes to be propagated to the caller
		if clone.RelativeBase != "" {
			clone.RelativeBase = normalizeBase(clone.RelativeBase)
		}
		// if the relative base is empty, let the schema loader choose a pseudo root document
		return &clone
	}
	return &specs.ExpandOptions{}
}

// ComposeSpec composes the references in a swagger spec
func ComposeSpec(spec *Swagger, options *specs.ExpandOptions) error {
	options = optionsOrDefault(options)
	resolver := defaultSchemaLoader(spec, options, nil, nil)

	specBasePath := options.RelativeBase

	if !options.SkipSchemas {
		for key, definition := range spec.Definitions {
			parentRefs := make([]string, 0, 10)
			parentRefs = append(parentRefs, fmt.Sprintf("#/definitions/%s", key))

			def, err := composeSchema(definition, parentRefs, resolver, specBasePath)
			if resolver.shouldStopOnError(err) {
				return err
			}
			if def != nil {
				spec.Definitions[key] = *def
			}
		}
	}

	for key := range spec.Parameters {
		parameter := spec.Parameters[key]
		if err := composeParameterOrResponse(&parameter, resolver, specBasePath); resolver.shouldStopOnError(err) {
			return err
		}
		spec.Parameters[key] = parameter
	}

	for key := range spec.Responses {
		response := spec.Responses[key]
		if err := composeParameterOrResponse(&response, resolver, specBasePath); resolver.shouldStopOnError(err) {
			return err
		}
		spec.Responses[key] = response
	}

	if spec.Paths != nil {
		for key := range spec.Paths.Paths {
			pth := spec.Paths.Paths[key]
			if err := composePathItem(&pth, resolver, specBasePath); resolver.shouldStopOnError(err) {
				return err
			}
			spec.Paths.Paths[key] = pth
		}
	}

	return nil
}

const rootBase = ".root"

// baseForRoot loads in the cache the root document and produces a fake ".root" base path entry
// for further $ref resolution
//
// Setting the cache is optional and this parameter may safely be left to nil.
func baseForRoot(root interface{}, cache specs.ResolutionCache) string {
	if root == nil {
		return ""
	}

	// cache the root document to resolve $ref's
	normalizedBase := normalizeBase(rootBase)
	cache.Set(normalizedBase, root)

	return normalizedBase
}

func composeItems(target specs.Schema, parentRefs []string, resolver *schemaLoader, basePath string) (*specs.Schema, error) {
	if target.Items == nil {
		return &target, nil
	}

	// array
	if target.Items.Schema != nil {
		t, err := composeSchema(*target.Items.Schema, parentRefs, resolver, basePath)
		if err != nil {
			return nil, err
		}
		*target.Items.Schema = *t
	}

	// tuple
	for i := range target.Items.Schemas {
		t, err := composeSchema(target.Items.Schemas[i], parentRefs, resolver, basePath)
		if err != nil {
			return nil, err
		}
		target.Items.Schemas[i] = *t
	}

	return &target, nil
}

func composeSchema(target specs.Schema, parentRefs []string, resolver *schemaLoader, basePath string) (*specs.Schema, error) {
	if target.Ref.String() == "" && target.Ref.IsRoot() {
		newRef := normalizeRef(&target.Ref, basePath)
		target.Ref = *newRef
		return &target, nil
	}

	// change the base path of resolution when an ID is encountered
	// otherwise the basePath should inherit the parent's
	if target.ID != "" {
		basePath, _ = resolver.setSchemaID(target, target.ID, basePath)
	}

	if target.Ref.String() != "" {
		return composeSchemaRef(target, parentRefs, resolver, basePath)
	}

	for k := range target.Definitions {
		tt, err := composeSchema(target.Definitions[k], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if tt != nil {
			target.Definitions[k] = *tt
		}
	}

	t, err := composeItems(target, parentRefs, resolver, basePath)
	if resolver.shouldStopOnError(err) {
		return &target, err
	}
	if t != nil {
		target = *t
	}

	for i := range target.AllOf {
		t, err := composeSchema(target.AllOf[i], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.AllOf[i] = *t
		}
	}

	for i := range target.AnyOf {
		t, err := composeSchema(target.AnyOf[i], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.AnyOf[i] = *t
		}
	}

	for i := range target.OneOf {
		t, err := composeSchema(target.OneOf[i], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.OneOf[i] = *t
		}
	}

	if target.Not != nil {
		t, err := composeSchema(*target.Not, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.Not = *t
		}
	}

	for k := range target.Properties {
		t, err := composeSchema(target.Properties[k], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.Properties[k] = *t
		}
	}

	if target.AdditionalProperties != nil && target.AdditionalProperties.Schema != nil {
		t, err := composeSchema(*target.AdditionalProperties.Schema, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.AdditionalProperties.Schema = *t
		}
	}

	for k := range target.PatternProperties {
		t, err := composeSchema(target.PatternProperties[k], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.PatternProperties[k] = *t
		}
	}

	for k := range target.Dependencies {
		if target.Dependencies[k].Schema != nil {
			t, err := composeSchema(*target.Dependencies[k].Schema, parentRefs, resolver, basePath)
			if resolver.shouldStopOnError(err) {
				return &target, err
			}
			if t != nil {
				*target.Dependencies[k].Schema = *t
			}
		}
	}

	if target.AdditionalItems != nil && target.AdditionalItems.Schema != nil {
		t, err := composeSchema(*target.AdditionalItems.Schema, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.AdditionalItems.Schema = *t
		}
	}
	return &target, nil
}

func composeSchemaRef(target specs.Schema, parentRefs []string, resolver *schemaLoader, basePath string) (*specs.Schema, error) {
	// if a Ref is found, all sibling fields are skipped
	// Ref also changes the resolution scope of children composeSchema

	// here the resolution scope is changed because a $ref was encountered
	normalizedRef := normalizeRef(&target.Ref, basePath)
	normalizedBasePath := normalizedRef.RemoteURI()

	if resolver.isCircular(normalizedRef, basePath, parentRefs...) {
		// this means there is a cycle in the recursion tree: return the Ref
		// - circular refs cannot be composeed. We leave them as ref.
		// - denormalization means that a new local file ref is set relative to the original basePath
		if !resolver.options.AbsoluteCircularRef {
			target.Ref = denormalizeRef(normalizedRef, resolver.context.basePath, resolver.context.rootID)
		} else {
			target.Ref = *normalizedRef
		}
		return &target, nil
	}

	var t *specs.Schema
	err := resolver.Resolve(&target.Ref, &t, basePath)
	if resolver.shouldStopOnError(err) {
		return nil, err
	}

	if t == nil {
		// guard for when continuing on error
		return &target, nil
	}

	parentRefs = append(parentRefs, normalizedRef.String())
	transitiveResolver := resolver.transitiveResolver(basePath, target.Ref)

	basePath = resolver.updateBasePath(transitiveResolver, normalizedBasePath)

	return composeSchema(*t, parentRefs, transitiveResolver, basePath)
}

func composePathItem(pathItem *PathItem, resolver *schemaLoader, basePath string) error {
	if pathItem == nil {
		return nil
	}

	parentRefs := make([]string, 0, 10)
	if err := resolver.deref(pathItem, parentRefs, basePath); resolver.shouldStopOnError(err) {
		return err
	}

	if pathItem.Ref.String() != "" {
		transitiveResolver := resolver.transitiveResolver(basePath, pathItem.Ref)
		basePath = transitiveResolver.updateBasePath(resolver, basePath)
		resolver = transitiveResolver
	}

	pathItem.Ref = specs.Ref{}
	for i := range pathItem.Parameters {
		if err := composeParameterOrResponse(&(pathItem.Parameters[i]), resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
	}

	ops := []*Operation{
		pathItem.Get,
		pathItem.Head,
		pathItem.Options,
		pathItem.Put,
		pathItem.Post,
		pathItem.Patch,
		pathItem.Delete,
	}
	for _, op := range ops {
		if err := composeOperation(op, resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
	}

	return nil
}

func composeContent(content interface{}, resolver *schemaLoader, basePath string) error {
	if content == nil {
		return nil
	}

	parentRefs := make([]string, 0, 10)
	if err := resolver.deref(content, parentRefs, basePath); resolver.shouldStopOnError(err) {
		return err
	}

	ref, sch, _ := getRefAndSchema(content)
	if ref.String() != "" {
		transitiveResolver := resolver.transitiveResolver(basePath, *ref)
		basePath = resolver.updateBasePath(transitiveResolver, basePath)
		resolver = transitiveResolver
	}

	if sch == nil {
		// nothing to be composeed
		if ref != nil {
			*ref = specs.Ref{}
		}
		return nil
	}

	if sch.Ref.String() != "" {
		rebasedRef, ern := specs.NewRef(normalizeURI(sch.Ref.String(), basePath))
		if ern != nil {
			return ern
		}

		switch {
		case resolver.isCircular(&rebasedRef, basePath, parentRefs...):
			// this is a circular $ref: stop expansion
			if !resolver.options.AbsoluteCircularRef {
				sch.Ref = denormalizeRef(&rebasedRef, resolver.context.basePath, resolver.context.rootID)
			} else {
				sch.Ref = rebasedRef
			}
		case !resolver.options.SkipSchemas:
			// schema composeed to a $ref in another root
			sch.Ref = rebasedRef
		default:
			// skip schema expansion but rebase $ref to schema
			sch.Ref = denormalizeRef(&rebasedRef, resolver.context.basePath, resolver.context.rootID)
		}
	}

	if ref != nil {
		*ref = specs.Ref{}
	}

	// compose schema
	if !resolver.options.SkipSchemas {
		s, err := composeSchema(*sch, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return err
		}
		if s == nil {
			// guard for when continuing on error
			return nil
		}
		*sch = *s
	}

	return nil
}

func composeOperation(op *Operation, resolver *schemaLoader, basePath string) error {
	if op == nil {
		return nil
	}

	for i := range op.Parameters {
		param := op.Parameters[i]
		if err := composeParameterOrResponse(&param, resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
		op.Parameters[i] = param
	}

	if op.Responses == nil {
		return nil
	}

	responses := op.Responses
	if err := composeParameterOrResponse(responses.Default, resolver, basePath); resolver.shouldStopOnError(err) {
		return err
	}

	for code := range responses.StatusCodeResponses {
		response := responses.StatusCodeResponses[code]
		if err := composeParameterOrResponse(&response, resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
		responses.StatusCodeResponses[code] = response

		for mediaType := range responses.StatusCodeResponses[code].Content {
			if responses.StatusCodeResponses[code].Content == nil {
				continue
			}
			content := responses.StatusCodeResponses[code].Content[mediaType]

			if err := composeContent(&content, resolver, basePath); resolver.shouldStopOnError(err) {
				return err
			}
			responses.StatusCodeResponses[code].Content[mediaType] = content
		}
	}

	return nil
}

func getRefAndSchema(input interface{}) (*specs.Ref, *specs.Schema, error) {
	var (
		ref *specs.Ref
		sch *specs.Schema
	)

	switch refable := input.(type) {
	case *specs.Parameter:
		if refable == nil {
			return nil, nil, nil
		}
		ref = &refable.Ref
		sch = refable.Schema
	case *Response:
		if refable == nil {
			return nil, nil, nil
		}
		ref = &refable.Ref
		sch = refable.Schema
	case *Content:
		if refable == nil {
			return nil, nil, nil
		}
		ref = &refable.Ref
		sch = refable.Schema
	default:
		return nil, nil, fmt.Errorf("unsupported type: %T: %w", input, ErrComposeUnsupportedType)
	}

	return ref, sch, nil
}

func composeParameterOrResponse(input interface{}, resolver *schemaLoader, basePath string) error {
	ref, _, err := getRefAndSchema(input)
	if err != nil {
		return err
	}

	if ref == nil {
		return nil
	}

	parentRefs := make([]string, 0, 10)
	if err = resolver.deref(input, parentRefs, basePath); resolver.shouldStopOnError(err) {
		return err
	}

	ref, sch, _ := getRefAndSchema(input)
	if ref.String() != "" {
		transitiveResolver := resolver.transitiveResolver(basePath, *ref)
		basePath = resolver.updateBasePath(transitiveResolver, basePath)
		resolver = transitiveResolver
	}

	if sch == nil {
		// nothing to be composeed
		if ref != nil {
			*ref = specs.Ref{}
		}
		return nil
	}

	if sch.Ref.String() != "" {
		rebasedRef, ern := specs.NewRef(normalizeURI(sch.Ref.String(), basePath))
		if ern != nil {
			return ern
		}

		switch {
		case resolver.isCircular(&rebasedRef, basePath, parentRefs...):
			// this is a circular $ref: stop expansion
			if !resolver.options.AbsoluteCircularRef {
				sch.Ref = denormalizeRef(&rebasedRef, resolver.context.basePath, resolver.context.rootID)
			} else {
				sch.Ref = rebasedRef
			}
		case !resolver.options.SkipSchemas:
			// schema composeed to a $ref in another root
			sch.Ref = rebasedRef
		default:
			// skip schema expansion but rebase $ref to schema
			sch.Ref = denormalizeRef(&rebasedRef, resolver.context.basePath, resolver.context.rootID)
		}
	}

	if ref != nil {
		*ref = specs.Ref{}
	}

	// compose schema
	if !resolver.options.SkipSchemas {
		s, err := composeSchema(*sch, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return err
		}
		if s == nil {
			// guard for when continuing on error
			return nil
		}
		*sch = *s
	}

	return nil
}
