//nolint:misspell // serialization package names and aliases are intentional.
package jsonschema

import (
	"context"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	jsonserialization "github.com/ARM-software/golang-utils/utils/serialization/json" //nolint:misspell
	"github.com/ARM-software/golang-utils/utils/serialization/yaml"                   //nolint:misspell
)

type fileValidator struct {
	schemaCreationFunc func(context.Context) (*jsonschema.Schema, error)
	expectedExtensions []string
	convertFunc        func([]byte) ([]byte, error)
}

func (c *fileValidator) ValidateContent(ctx context.Context, a any) error {
	if bytes, ok := a.([]byte); ok {
		if c.convertFunc != nil {
			var err error
			bytes, err = c.convertFunc(bytes)
			if err != nil {
				return err
			}
		}
		var decoded any
		err := jsonserialization.UnmarshallWithContext(ctx, bytes, &decoded)
		if err != nil {
			return commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to decode JSON content")
		}
		a = decoded
	}
	schema, err := c.schemaCreationFunc(ctx)
	if err != nil {
		return err
	}
	return validateAgainstSchema(schema, a)
}

func (c *fileValidator) ValidateFile(ctx context.Context, filepath string) error {
	return c.ValidateFileInFS(ctx, filesystem.GetGlobalFileSystem(), filepath)
}

func (c *fileValidator) ValidateFileInFS(ctx context.Context, fs filesystem.FS, filepath string) (err error) {
	if fs == nil {
		err = commonerrors.UndefinedVariable("file system")
		return
	}
	err = filesystem.NewPathExtensionRule(fs, true, c.expectedExtensions...).Validate(filepath)
	if err != nil {
		return
	}
	err = filesystem.NewPathExistRule(fs, true).Validate(filepath)
	if err != nil {
		return
	}
	content, err := fs.ReadFileWithContext(ctx, filepath)
	if err != nil {
		return
	}
	err = c.ValidateContent(ctx, content)
	return
}

func newSchemaCreationFunc(schemaID *string, schema ...Schema) func(context.Context) (*jsonschema.Schema, error) {
	var once sync.Once
	var compiled *jsonschema.Schema
	var compiledErr error

	return func(ctx context.Context) (*jsonschema.Schema, error) {
		once.Do(func() {
			compileCtx := context.Background()
			if ctx != nil {
				compileCtx = context.WithoutCancel(ctx)
			}
			compiled, compiledErr = generateSchemaDefinition(compileCtx, schemaID, schema...)
		})
		return compiled, compiledErr
	}
}

// NewJSONFileValidator returns a validator for JSON content and `.json` files.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
//
// The returned validator compiles its schema set lazily and caches the result
// for reuse across subsequent validations.
func NewJSONFileValidator(schemaID *string, schema ...Schema) (v ISchemaValidator, err error) {
	v = &fileValidator{
		schemaCreationFunc: newSchemaCreationFunc(schemaID, schema...),
		expectedExtensions: []string{".json"},
		convertFunc:        nil,
	}
	return
}

// NewJSONFileValidatorWithOptions returns a JSON validator using a Schema built
// from the supplied options.
func NewJSONFileValidatorWithOptions(options ...SchemaOption) (ISchemaValidator, error) {
	return NewJSONFileValidator(nil, *NewJSONSchemaFile(options...))
}

// NewYAMLFileValidator returns a validator for YAML content and `.yaml` or
// `.yml` files.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
//
// The returned validator compiles its schema set lazily and caches the result
// for reuse across subsequent validations.
func NewYAMLFileValidator(schemaID *string, schema ...Schema) (v ISchemaValidator, err error) {
	v = &fileValidator{
		schemaCreationFunc: newSchemaCreationFunc(schemaID, schema...),
		expectedExtensions: []string{".yaml", ".yml"},
		convertFunc:        yaml.ToJSON,
	}
	return
}

// NewYAMLFileValidatorWithOptions returns a YAML validator using a Schema built
// from the supplied options.
func NewYAMLFileValidatorWithOptions(options ...SchemaOption) (ISchemaValidator, error) {
	return NewYAMLFileValidator(nil, *NewJSONSchemaFile(options...))
}
