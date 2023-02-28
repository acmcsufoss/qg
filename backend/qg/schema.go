package qg

import (
	"encoding/json"
	"fmt"
	"strings"

	_ "embed"

	"github.com/davecgh/go-spew/spew"
	jtd "github.com/jsontypedef/json-typedef-go"
	"github.com/pkg/errors"
)

// Schema is the JSON Typedef schema for the qg package.
var Schema jtd.Schema

//go:embed schema.json
var schemaJSON []byte

func init() {
	if err := json.Unmarshal(schemaJSON, &Schema); err != nil {
		panic("cannot unmarshal qg schema: " + err.Error())
	}
}

// // Unmarshal unmarshals the given value into a new instance.
// func Unmarshal[T any](b []byte) (*T, error) {
// 	var z T
// 	k := fmt.Sprintf("%T", z)
// 	panic(k)
// }

// func

// TypeValidationError is an error that occurs when a value does not
// match the expected type definition.
type TypeValidationError jtd.ValidateError

func (err TypeValidationError) Error() string {
	path := "$"
	if len(err.InstancePath) > 0 {
		path += "." + strings.Join(err.InstancePath, ".")
	}

	return fmt.Sprintf(
		"error at %s (%s)",
		path, strings.Join(err.SchemaPath, "."))
}

// Validate validates the given value against the qg schema.
func Validate(name string, v any) error {
	vschema, ok := Schema.Definitions[name]
	if !ok {
		return fmt.Errorf("cannot find schema %q", name)
	}
	vschema.Definitions = Schema.Definitions

	// Dogshit code to make the validator happy.
	m, err := asJSONMap(v)
	if err != nil {
		return errors.Wrap(err, "marshaling value to JSON")
	}

	spew.Dump(m)

	errors, err := jtd.Validate(vschema, m,
		jtd.WithMaxDepth(100),
		jtd.WithMaxErrors(1))
	if err != nil {
		return fmt.Errorf("cannot validate %q: %w", name, err)
	}

	if len(errors) == 0 {
		return nil
	}

	return TypeValidationError(errors[0])
}

func asJSONMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return m, nil
}
