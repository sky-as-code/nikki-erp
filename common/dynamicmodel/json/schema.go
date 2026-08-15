// Package json holds the JSON Schema describing a dynamic model definition,
// plus the validator used before a model JSON string is parsed into a builder.
//
// This package must not import common/dynamicmodel/model: the dependency runs the
// other way, so that model_builder_json.go can validate before parsing.
package json

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"go.bryk.io/pkg/errors"
)

// schemaResourceUrl is the in-memory URL the schema document is registered under.
// It never leaves the process; the compiler only needs a stable key.
const schemaResourceUrl = "https://sky-as-code.github.io/nikki-erp/schemas/model_schema.json"

// ModelJsonSchema is the JSON Schema every model JSON file is validated against.
//
//go:embed model_schema.json
var ModelJsonSchema string

var compiledCache = struct {
	schemas map[string]*jsonschema.Schema
	mu      *sync.RWMutex
}{
	schemas: make(map[string]*jsonschema.Schema),
	mu:      &sync.RWMutex{},
}

// compileSchema compiles a JSON Schema document, caching the result by its source text.
// Compiling is expensive and every model in a module shares the same schema, so the
// cache turns N compilations into one.
func compileSchema(schemaJson string) (*jsonschema.Schema, error) {
	compiledCache.mu.RLock()
	cached, exists := compiledCache.schemas[schemaJson]
	compiledCache.mu.RUnlock()
	if exists {
		return cached, nil
	}

	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaJson))
	if err != nil {
		return nil, errors.Wrap(err, "compileSchema: JSON Schema is not valid JSON")
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaResourceUrl, doc); err != nil {
		return nil, errors.Wrap(err, "compileSchema: cannot add schema resource")
	}
	compiled, err := compiler.Compile(schemaResourceUrl)
	if err != nil {
		return nil, errors.Wrap(err, "compileSchema: cannot compile JSON Schema")
	}

	compiledCache.mu.Lock()
	compiledCache.schemas[schemaJson] = compiled
	compiledCache.mu.Unlock()

	return compiled, nil
}
