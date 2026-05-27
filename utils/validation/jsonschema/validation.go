// Package jsonschema provides helpers for validating JSON and YAML content
// against JSON Schema documents.
//
// The package wraps github.com/santhosh-tekuri/jsonschema/v6 and standardises
// a few behaviours used across this repository:
//   - schema definitions can be described as files and loaded from either the
//     normal filesystem or an embedded filesystem
//   - raw JSON values, raw YAML bytes, JSON files, and YAML files can all be
//     validated through the same schema abstraction
//   - validation and loading failures are wrapped with the repository's common
//     error types so callers can test and report them consistently
//   - when a schema is composed of multiple schema documents, callers may pass
//     an optional schema ID to select the root schema to compile; otherwise it
//     should be left as nil
//
// In normal use, callers describe one or more schema files with Schema values,
// optionally pass a schema ID when the schema set is composed of multiple
// documents, and then validate either raw content or files through the helper
// functions in this package.
package jsonschema

import (
	"context"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	santhoshjsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/serialization/json" //nolint:misspell
	"github.com/ARM-software/golang-utils/utils/serialization/yaml" //nolint:misspell
)

// Schema describes a schema file that can be loaded and registered for
// validation.
//
// It is intended to be a lightweight wrapper around a schema file so callers
// can build validators from normal or embedded filesystems without passing raw
// schema bytes around manually.
type Schema struct {
	// Title is a human-readable title of the schema.
	Title string
	// LocalPath is the path to the schema file. It should be available on the
	// filesystem or an embedded filesystem.
	LocalPath string
	// ID should match the $id field within the schema.
	ID string
	// Filesystem is the filesystem to use to load the schema file.
	Filesystem filesystem.FS

	// Limits defines the filesystem read limits used when loading the schema
	// file.
	Limits filesystem.ILimits
}

// Validate checks that the schema definition contains the required fields
// needed to load a schema file.
func (s *Schema) Validate() error {
	if s == nil {
		return commonerrors.UndefinedVariable("schema")
	}
	err := config.ValidateEmbedded(s)
	if err == nil {
		err = validation.ValidateStruct(s,
			validation.Field(&s.Title, validation.Required),
			validation.Field(&s.LocalPath, validation.Required),
			validation.Field(&s.ID, validation.Required),
			validation.Field(&s.Filesystem, validation.Required),
			validation.Field(&s.Limits, validation.Required),
		)
	}

	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid json schema definition")
	}
	return nil
}

// SchemaSpec contains the raw JSON schema document and the identifier used to
// register it with the JSON Schema compiler.
type SchemaSpec struct {
	ID            string
	Specification []byte
	// Title is a human-readable title of the schema.
	Title string
}

// Validate checks that the schema specification contains the required fields
// needed to register and compile a schema.
func (s *SchemaSpec) Validate() error {
	if s == nil {
		return commonerrors.UndefinedVariable("schema specification")
	}
	err := validation.ValidateStruct(s,
		validation.Field(&s.ID, validation.Required),
		validation.Field(&s.Specification, validation.Required),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid json schema specification")
	}
	return nil
}

// LoadSchemaSpec reads and returns the raw schema specification described by
// schema.
//
// The schema file is read from the filesystem attached to the Schema, which may
// be a normal or embedded filesystem.
func LoadSchemaSpec(ctx context.Context, schema *Schema) (*SchemaSpec, error) {
	if schema == nil {
		return nil, commonerrors.UndefinedVariable("schema")
	}
	schema = schema.Default()
	err := schema.Validate()
	if err != nil {
		return nil, err
	}
	schemaPath := filesystem.FilePathClean(schema.Filesystem, schema.LocalPath)

	data, err := schema.Filesystem.ReadFileWithContextAndLimits(ctx, schemaPath, schema.Limits)
	if err != nil {
		return nil, commonerrors.DescribeCircumstancef(err, "failed to load JSON Schema [%v] from [%v]", schema.Title, schema.LocalPath)
	}
	if yaml.IsYAMLFile(filesystem.FilePathExt(schema.Filesystem, schemaPath)) {
		data, err = yaml.ToJSON(data)
		if err != nil {
			return nil, commonerrors.DescribeCircumstancef(err, "failed to convert schema from YAML to JSON [%v]", schema.Title)
		}
	}

	return &SchemaSpec{
		ID:            schemaID(schema.ID, schema.LocalPath),
		Specification: data,
		Title:         schema.Title,
	}, nil
}

// ValidateRawJSONAgainstSchema validates JSON-compatible content against the
// supplied schema definitions.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
//
// The content may already be decoded into Go values, or any other value that
// can be passed directly to the underlying JSON Schema validator.
func ValidateRawJSONAgainstSchema(ctx context.Context, content any, schemaID *string, schema ...Schema) error {
	v, err := NewJSONFileValidator(schemaID, schema...)
	if err != nil {
		return err
	}
	return v.ValidateContent(ctx, content)
}

// ValidateRawJSONAgainstSchemaOptions validates JSON-compatible content against
// a Schema built from the supplied options.
func ValidateRawJSONAgainstSchemaOptions(ctx context.Context, content any, options ...SchemaOption) error {
	return ValidateRawJSONAgainstSchema(ctx, content, nil, *NewJSONSchemaFile(options...))
}

// ValidateRawYAMLAgainstSchema validates raw YAML bytes against the supplied
// schema definitions.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
//
// The YAML content is converted to JSON first and then validated against the
// compiled JSON Schema.
func ValidateRawYAMLAgainstSchema(ctx context.Context, content []byte, schemaID *string, schema ...Schema) error {
	v, err := NewYAMLFileValidator(schemaID, schema...)
	if err != nil {
		return err
	}
	return v.ValidateContent(ctx, content)
}

// ValidateRawYAMLAgainstSchemaOptions validates raw YAML bytes against a Schema
// built from the supplied options.
func ValidateRawYAMLAgainstSchemaOptions(ctx context.Context, content []byte, options ...SchemaOption) error {
	return ValidateRawYAMLAgainstSchema(ctx, content, nil, *NewJSONSchemaFile(options...))
}

// ValidateJSONFileAgainstSchemaFS validates a JSON file from the supplied
// filesystem against the supplied schema definitions.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
func ValidateJSONFileAgainstSchemaFS(ctx context.Context, fileSystem filesystem.FS, filePath string, schemaID *string, schema ...Schema) error {
	v, err := NewJSONFileValidator(schemaID, schema...)
	if err != nil {
		return err
	}
	return v.ValidateFileInFS(ctx, fileSystem, filePath)
}

// ValidateJSONFileAgainstSchemaFSWithOptions validates a JSON file from the
// supplied filesystem against a Schema built from the supplied options.
func ValidateJSONFileAgainstSchemaFSWithOptions(ctx context.Context, fileSystem filesystem.FS, filePath string, options ...SchemaOption) error {
	return ValidateJSONFileAgainstSchemaFS(ctx, fileSystem, filePath, nil, *NewJSONSchemaFile(options...))
}

// ValidateYAMLFileAgainstSchemaFS validates a YAML file from the supplied
// filesystem against the supplied schema definitions.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
func ValidateYAMLFileAgainstSchemaFS(ctx context.Context, fileSystem filesystem.FS, filePath string, schemaID *string, schema ...Schema) error {
	v, err := NewYAMLFileValidator(schemaID, schema...)
	if err != nil {
		return err
	}
	return v.ValidateFileInFS(ctx, fileSystem, filePath)
}

// ValidateYAMLFileAgainstSchemaFSWithOptions validates a YAML file from the
// supplied filesystem against a Schema built from the supplied options.
func ValidateYAMLFileAgainstSchemaFSWithOptions(ctx context.Context, fileSystem filesystem.FS, filePath string, options ...SchemaOption) error {
	return ValidateYAMLFileAgainstSchemaFS(ctx, fileSystem, filePath, nil, *NewJSONSchemaFile(options...))
}

// ValidateJSONFileAgainstSchema validates a JSON file from the global
// filesystem against the supplied schema definitions.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
func ValidateJSONFileAgainstSchema(ctx context.Context, filePath string, schemaID *string, schema ...Schema) error {
	v, err := NewJSONFileValidator(schemaID, schema...)
	if err != nil {
		return err
	}
	return v.ValidateFile(ctx, filePath)
}

// ValidateJSONFileAgainstSchemaOptions validates a JSON file against a Schema
// built from the supplied options.
func ValidateJSONFileAgainstSchemaOptions(ctx context.Context, filePath string, options ...SchemaOption) error {
	return ValidateJSONFileAgainstSchema(ctx, filePath, nil, *NewJSONSchemaFile(options...))
}

// ValidateYAMLFileAgainstSchema validates a YAML file from the global
// filesystem against the supplied schema definitions.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
func ValidateYAMLFileAgainstSchema(ctx context.Context, filePath string, schemaID *string, schema ...Schema) error {
	v, err := NewYAMLFileValidator(schemaID, schema...)
	if err != nil {
		return err
	}
	return v.ValidateFile(ctx, filePath)
}

// ValidateYAMLFileAgainstSchemaOptions validates a YAML file against a Schema
// built from the supplied options.
func ValidateYAMLFileAgainstSchemaOptions(ctx context.Context, filePath string, options ...SchemaOption) error {
	return ValidateYAMLFileAgainstSchema(ctx, filePath, nil, *NewJSONSchemaFile(options...))
}

// validateAgainstSchema validates a file against a schema
func validateAgainstSchema(schema *santhoshjsonschema.Schema, content any) (err error) {
	if reflection.IsEmpty(content) {
		err = commonerrors.UndefinedVariable("content to validate against schema")
		return
	}
	if reflection.IsEmpty(schema) {
		err = commonerrors.UndefinedVariable("schema to validate against")
		return
	}
	err = schema.Validate(content)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "content does not comply with JSON schema")
		return
	}
	return
}

// generateSchemaDefinition compiles the schema set used for validation.
//
// `schemaID` is the optional schema ID to compile when the schema is composed
// of multiple schema documents. Otherwise, leave it as nil.
func generateSchemaDefinition(ctx context.Context, schemaID *string, composingSchemas ...Schema) (schema *santhoshjsonschema.Schema, err error) {
	if len(composingSchemas) == 0 {
		err = commonerrors.UndefinedVariable("schemas")
		return
	}
	id := field.OptionalString(schemaID, composingSchemas[0].ID)

	compiler := santhoshjsonschema.NewCompiler()
	err = collection.EachRef[Schema](slices.Values(composingSchemas), func(schema *Schema) error {
		spec, subErr := LoadSchemaSpec(ctx, schema)
		if subErr != nil {
			return subErr
		}
		return registerSchemaToCompiler(ctx, compiler, spec)
	})
	if err != nil {
		return nil, err
	}

	schema, err = compiler.Compile(id)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "failed to compile schema [%v]", id)
	}
	return
}

func registerSchemaToCompiler(ctx context.Context, compiler *santhoshjsonschema.Compiler, spec *SchemaSpec) (err error) {
	if compiler == nil {
		err = commonerrors.UndefinedVariable("json schema compiler")
		return
	}
	if spec == nil {
		err = commonerrors.UndefinedVariable("schema specification")
		return
	}
	err = spec.Validate()
	if err != nil {
		return
	}

	var doc any
	err = json.UnmarshallWithContext(ctx, spec.Specification, &doc)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrMarshalling, err, "failed to decode JSON Schema specification [%v]", spec.Title)
		return
	}

	err = compiler.AddResource(spec.ID, doc)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "failed to register schema [%v] to schema compiler", spec.Title)
	}
	return
}

func schemaID(id string, localPath string) string {
	id = strings.TrimSpace(id)
	if !reflection.IsEmpty(id) {
		return id
	}
	return strings.TrimSpace(localPath)
}
