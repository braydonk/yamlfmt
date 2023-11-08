package basic

import (
	"errors"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/braydonk/yaml"
	"github.com/google/yamlfmt/internal/collections"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var ErrSchemaNoPathMatch = errors.New("path failed to match")

// YAMLSchema is a representation of a JSON schema to use to
// validate against any path that matches the match path.
type YAMLSchema struct {
	URL   string `mapstructure:"url"`
	Match string `mapstructure:"match"`

	schema *jsonschema.Schema
}

// Validate will take in a path and data read from the file to
// validate against the schema if the path matches.
func (s *YAMLSchema) Validate(path string, data []byte) (bool, error) {
	if !s.matchPath(path) {
		return false, ErrSchemaNoPathMatch
	}
	// The schema will only be compiled once we know for sure a path
	// matches and we need to validate. If no paths ever match this
	// configured schema, then the schema will never be compiled.
	if s.schema == nil {
		err := s.compile()
		if err != nil {
			return false, err
		}
	}

	var v interface{}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return false, err
	}
	err := s.schema.Validate(data)
	if err != nil {
		fmt.Printf("Schema validation failed for %s:\n%v\n", path, err)
		return false, nil
	}
	return true, nil
}

func (s *YAMLSchema) matchPath(path string) bool {
	matched, err := doublestar.Match(s.Match, path)
	if err != nil {
		fmt.Printf("path match error: %v\n", err)
		return false
	}
	return matched
}

func (s *YAMLSchema) compile() error {
	schema, err := jsonschema.Compile(s.URL)
	if err != nil {
		return err
	}
	s.schema = schema
	return err
}

// Schemas is a registry of YAMLSchema structs that are generally part
// of one configuration.
type YAMLSchemas collections.Set[YAMLSchema]

// Validate will run the path and data through the Validate function on every schema.
//
// If there is a path match and an error in processing, it will return false and that
// error.
// If there is a path match and no error, it will return whether the schema
// validated or not.
// If there is no path match throughout the collection of schemas, it will return
// a ErrSchemaNoPathMatch error.
func (ys YAMLSchemas) Validate(path string, data []byte) (bool, error) {
	for schema := range ys {
		valid, err := schema.Validate(path, data)
		if err == nil {
			return valid, nil
		}
		if err != nil && !errors.Is(err, ErrSchemaNoPathMatch) {
			return false, err
		}
	}
	return false, ErrSchemaNoPathMatch
}
