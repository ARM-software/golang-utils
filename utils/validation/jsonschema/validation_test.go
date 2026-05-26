package jsonschema

import (
	"context"
	"embed"
	"path"
	"testing"

	santhoshjsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

const compoundRootSchemaID = "https://example.com/schemas/compound/root.schema.json"

var (
	//go:embed testdata/*.json testdata/*.yaml
	embeddedFS embed.FS
)

func newValidSchema(t *testing.T, fs filesystem.FS) *Schema {
	t.Helper()
	return NewJSONSchemaFile(
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithID("https://example.com/schemas/person.schema.json"),
		WithFilesystem(fs),
	)
}

func newSchema(t *testing.T, fs filesystem.FS, title string, localPath string, id string) *Schema {
	t.Helper()
	return NewJSONSchemaFile(WithTitle(title), WithLocalPath(localPath), WithID(id), WithFilesystem(fs))
}

func newMissingSchema(t *testing.T, fs filesystem.FS) *Schema {
	t.Helper()
	return newSchema(
		t,
		fs,
		"missing schema",
		path.Join("testdata", "missing.schema.json"),
		"https://example.com/schemas/missing.schema.json",
	)
}

func newEmbeddedFilesystem(t *testing.T) filesystem.FS {
	t.Helper()
	fs, err := filesystem.NewEmbedFileSystem(&embeddedFS)
	require.NoError(t, err)
	return fs
}

func newCompoundSchemas(t *testing.T, fs filesystem.FS) []Schema {
	t.Helper()
	return []Schema{
		{
			Title:      "compound name",
			LocalPath:  path.Join("testdata", "compound_name.schema.json"),
			ID:         "https://example.com/schemas/compound/name.schema.json",
			Filesystem: fs,
		},
		{
			Title:      "compound count",
			LocalPath:  path.Join("testdata", "compound_count.schema.json"),
			ID:         "https://example.com/schemas/compound/count.schema.json",
			Filesystem: fs,
		},
		{
			Title:      "compound root",
			LocalPath:  path.Join("testdata", "compound_root.schema.json"),
			ID:         compoundRootSchemaID,
			Filesystem: fs,
		},
	}
}

func TestSchemaValidate(t *testing.T) {
	require.NoError(t, newValidSchema(t, filesystem.GetGlobalFileSystem()).Validate())

	var nilSchema *Schema
	err := nilSchema.Validate()
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrUndefined)

	err = (&Schema{}).Validate()
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestNewJSONSchemaFileDefaults(t *testing.T) {
	schema := NewJSONSchemaFile(WithTitle("person"), WithLocalPath(path.Join("testdata", "person.schema.json")))
	require.NotNil(t, schema)
	assert.Equal(t, "person", schema.Title)
	assert.Equal(t, path.Join("testdata", "person.schema.json"), schema.LocalPath)
	assert.Equal(t, path.Join("testdata", "person.schema.json"), schema.ID)
	assert.Equal(t, filesystem.GetGlobalFileSystem(), schema.Filesystem)
}

func TestNewJSONSchemaFile(t *testing.T) {
	schema := NewJSONSchemaFile(
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
	)
	require.NotNil(t, schema)
	assert.Equal(t, "person", schema.Title)
	assert.Equal(t, path.Join("testdata", "person.schema.json"), schema.ID)
}

func TestWithFileLimits(t *testing.T) {
	limits := filesystem.DefaultLimits()
	schema := NewJSONSchemaFile(
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFileLimits(limits),
	)
	require.NotNil(t, schema)
	assert.Equal(t, limits, schema.Limits)
}

func TestSchemaSpecValidate(t *testing.T) {
	require.NoError(t, (&SchemaSpec{ID: "schema.json", Specification: []byte(`{"type":"object"}`)}).Validate())

	var nilSpec *SchemaSpec
	err := nilSpec.Validate()
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrUndefined)

	err = (&SchemaSpec{}).Validate()
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestLoadSchemaSpec(t *testing.T) {
	spec, err := LoadSchemaSpec(context.Background(), newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/schemas/person.schema.json", spec.ID)
	assert.NotEmpty(t, spec.Specification)
}

func TestLoadSchemaSpecEmbeddedFS(t *testing.T) {
	spec, err := LoadSchemaSpec(context.Background(), newValidSchema(t, newEmbeddedFilesystem(t)))
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/schemas/person.schema.json", spec.ID)
	assert.NotEmpty(t, spec.Specification)
}

func TestLoadSchemaSpec_MissingSchemaPath(t *testing.T) {
	_, err := LoadSchemaSpec(context.Background(), newMissingSchema(t, filesystem.GetGlobalFileSystem()))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.ErrorContains(t, err, "failed to load JSON Schema")
}

func TestValidateRawJSONAgainstSchema(t *testing.T) {
	err := ValidateRawJSONAgainstSchema(context.Background(), map[string]any{"name": "Alice", "count": 2}, nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
}

func TestValidateRawJSONAgainstSchemaOptions(t *testing.T) {
	err := ValidateRawJSONAgainstSchemaOptions(
		context.Background(),
		map[string]any{"name": "Alice", "count": 2},
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
	)
	require.NoError(t, err)
}

func TestValidateRawJSONAgainstSchema_InvalidContent(t *testing.T) {
	err := ValidateRawJSONAgainstSchema(context.Background(), map[string]any{"count": "two"}, nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateRawJSONAgainstCompoundSchema(t *testing.T) {
	schemaID := compoundRootSchemaID
	err := ValidateRawJSONAgainstSchema(
		context.Background(),
		map[string]any{"name": "Alice", "count": 2},
		&schemaID,
		newCompoundSchemas(t, filesystem.GetGlobalFileSystem())...,
	)
	require.NoError(t, err)
}

func TestValidateGoogleDraft202012DefsFixture(t *testing.T) {
	// Source: https://github.com/json-schema-org/JSON-Schema-Test-Suite/blob/main/tests/draft2020-12/defs.json
	googleSchema := newSchema(
		t,
		filesystem.GetGlobalFileSystem(),
		"google defs metaschema fixture",
		path.Join("testdata", "google_defs_metaschema.schema.json"),
		"https://example.com/schemas/google/defs-metaschema.schema.json",
	)

	err := ValidateJSONFileAgainstSchema(
		context.Background(),
		path.Join("testdata", "google_defs_valid.json"),
		nil,
		*googleSchema,
	)
	require.NoError(t, err)

	err = ValidateJSONFileAgainstSchema(
		context.Background(),
		path.Join("testdata", "google_defs_invalid.json"),
		nil,
		*googleSchema,
	)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateGojsonschemaIntegerFixture(t *testing.T) {
	// Source: https://github.com/xeipuuv/gojsonschema/blob/master/testdata/remotes/integer.json
	gojsonschemaFixture := newSchema(
		t,
		filesystem.GetGlobalFileSystem(),
		"gojsonschema integer fixture",
		path.Join("testdata", "gojsonschema_integer.schema.json"),
		"https://example.com/schemas/gojsonschema/integer.json",
	)

	err := ValidateRawJSONAgainstSchema(context.Background(), 2, nil, *gojsonschemaFixture)
	require.NoError(t, err)

	err = ValidateRawJSONAgainstSchema(context.Background(), "2", nil, *gojsonschemaFixture)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateJSONSchemaTestSuiteRequiredFixture(t *testing.T) {
	// Source: https://github.com/json-schema-org/JSON-Schema-Test-Suite/blob/main/tests/draft2020-12/required.json
	requiredFixture := newSchema(
		t,
		filesystem.GetGlobalFileSystem(),
		"json-schema-test-suite required fixture",
		path.Join("testdata", "json_schema_test_suite_required.schema.json"),
		"https://example.com/schemas/json-schema-test-suite/required.schema.json",
	)

	err := ValidateJSONFileAgainstSchema(
		context.Background(),
		path.Join("testdata", "json_schema_test_suite_required_valid.json"),
		nil,
		*requiredFixture,
	)
	require.NoError(t, err)

	err = ValidateJSONFileAgainstSchema(
		context.Background(),
		path.Join("testdata", "json_schema_test_suite_required_invalid.json"),
		nil,
		*requiredFixture,
	)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateJSONSchemaTestSuiteBooleanFalseFixture(t *testing.T) {
	// Source: https://github.com/json-schema-org/JSON-Schema-Test-Suite/blob/main/tests/draft2020-12/boolean_schema.json
	booleanFalseFixture := newSchema(
		t,
		filesystem.GetGlobalFileSystem(),
		"json-schema-test-suite boolean false fixture",
		path.Join("testdata", "json_schema_test_suite_boolean_false.schema.json"),
		"https://example.com/schemas/json-schema-test-suite/boolean-false.schema.json",
	)

	err := ValidateRawJSONAgainstSchema(context.Background(), 1, nil, *booleanFalseFixture)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	err = ValidateRawJSONAgainstSchema(context.Background(), map[string]any{"foo": "bar"}, nil, *booleanFalseFixture)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateGojsonschemaFragmentSchemaFixture(t *testing.T) {
	// Source: https://github.com/xeipuuv/gojsonschema/blob/master/testdata/extra/fragment_schema.json
	fragmentFixture := newSchema(
		t,
		filesystem.GetGlobalFileSystem(),
		"gojsonschema fragment schema fixture",
		path.Join("testdata", "gojsonschema_fragment_schema.json"),
		"https://example.com/schemas/gojsonschema/fragment_schema.json",
	)

	schemaID := "https://example.com/schemas/gojsonschema/fragment_schema.json#/definitions/x"
	err := ValidateRawJSONAgainstSchema(context.Background(), 2, &schemaID, *fragmentFixture)
	require.NoError(t, err)

	err = ValidateRawJSONAgainstSchema(context.Background(), "2", &schemaID, *fragmentFixture)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateGojsonschemaSubSchemasFixture(t *testing.T) {
	// Source: https://github.com/xeipuuv/gojsonschema/blob/master/testdata/remotes/subSchemas.json
	subSchemasFixture := newSchema(
		t,
		filesystem.GetGlobalFileSystem(),
		"gojsonschema subSchemas fixture",
		path.Join("testdata", "gojsonschema_subSchemas.json"),
		"https://example.com/schemas/gojsonschema/subSchemas.json",
	)

	schemaID := "https://example.com/schemas/gojsonschema/subSchemas.json#/refToInteger"
	err := ValidateRawJSONAgainstSchema(context.Background(), 2, &schemaID, *subSchemasFixture)
	require.NoError(t, err)

	err = ValidateRawJSONAgainstSchema(context.Background(), "2", &schemaID, *subSchemasFixture)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateRawYAMLAgainstSchema(t *testing.T) {
	err := ValidateRawYAMLAgainstSchema(context.Background(), []byte("name: Alice\ncount: 2\n"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
}

func TestValidateRawYAMLAgainstSchemaOptions(t *testing.T) {
	err := ValidateRawYAMLAgainstSchemaOptions(
		context.Background(),
		[]byte("name: Alice\ncount: 2\n"),
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
	)
	require.NoError(t, err)
}

func TestValidateRawYAMLAgainstSchema_InvalidContent(t *testing.T) {
	err := ValidateRawYAMLAgainstSchema(context.Background(), []byte("count: two\n"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestRegisterSchemaToCompiler_InvalidSchemaSpec(t *testing.T) {
	err := registerSchemaToCompiler(context.Background(), santhoshjsonschema.NewCompiler(), &SchemaSpec{})
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestRegisterSchemaToCompiler_InvalidSpecificationJSON(t *testing.T) {
	err := registerSchemaToCompiler(context.Background(), santhoshjsonschema.NewCompiler(), &SchemaSpec{
		ID:            "schema.json",
		Title:         "broken",
		Specification: []byte(`{"type":`),
	})
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrMarshalling)
}

func TestGenerateSchemaDefinition(t *testing.T) {
	sch, err := generateSchemaDefinition(context.Background(), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
	assert.NotNil(t, sch)
}

func TestSchemaGenerationOnlyRunsOnceAfterCancelledContext(t *testing.T) {
	v, err := NewJSONFileValidator(nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, v.ValidateContent(cancelledCtx, map[string]any{"name": "Alice", "count": 2}))
	require.NoError(t, v.ValidateContent(context.Background(), map[string]any{"name": "Alice", "count": 2}))
}

func TestNewJSONFileValidatorWithOptions(t *testing.T) {
	v, err := NewJSONFileValidatorWithOptions(
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
	)
	require.NoError(t, err)
	require.NoError(t, v.ValidateContent(context.Background(), map[string]any{"name": "Alice", "count": 2}))
}

func TestNewYAMLFileValidatorWithOptions(t *testing.T) {
	v, err := NewYAMLFileValidatorWithOptions(
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
	)
	require.NoError(t, err)
	require.NoError(t, v.ValidateContent(context.Background(), []byte("name: Alice\ncount: 2\n")))
}

func TestNewJSONFileValidatorWithOptions_LowSchemaFileLimit(t *testing.T) {
	v, err := NewJSONFileValidatorWithOptions(
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
		WithFileLimits(filesystem.NewLimits(1, 1024, 1, 1, false)),
	)
	require.NoError(t, err)

	err = v.ValidateContent(context.Background(), map[string]any{"name": "Alice", "count": 2})
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTooLarge)
}

func TestValidateJSONFileAgainstSchema(t *testing.T) {
	err := ValidateJSONFileAgainstSchema(context.Background(), path.Join("testdata", "valid.json"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
}

func TestValidateJSONFileAgainstSchemaOptions(t *testing.T) {
	err := ValidateJSONFileAgainstSchemaOptions(
		context.Background(),
		path.Join("testdata", "valid.json"),
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
	)
	require.NoError(t, err)
}

func TestValidateJSONFileAgainstSchema_MissingSchemaPath(t *testing.T) {
	err := ValidateJSONFileAgainstSchema(context.Background(), path.Join("testdata", "valid.json"), nil, *newMissingSchema(t, filesystem.GetGlobalFileSystem()))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.ErrorContains(t, err, "failed to load JSON Schema")
}

func TestValidateJSONFileAgainstSchemaFS(t *testing.T) {
	err := ValidateJSONFileAgainstSchemaFS(context.Background(), newEmbeddedFilesystem(t), path.Join("testdata", "valid.json"), nil, *newValidSchema(t, newEmbeddedFilesystem(t)))
	require.NoError(t, err)
}

func TestValidateJSONFileAgainstSchemaFSWithOptions(t *testing.T) {
	err := ValidateJSONFileAgainstSchemaFSWithOptions(
		context.Background(),
		newEmbeddedFilesystem(t),
		path.Join("testdata", "valid.json"),
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(newEmbeddedFilesystem(t)),
	)
	require.NoError(t, err)
}

func TestValidateYAMLFileAgainstSchema(t *testing.T) {
	err := ValidateYAMLFileAgainstSchema(context.Background(), path.Join("testdata", "valid.yaml"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
}

func TestValidateYAMLFileAgainstSchemaOptions(t *testing.T) {
	err := ValidateYAMLFileAgainstSchemaOptions(
		context.Background(),
		path.Join("testdata", "valid.yaml"),
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(filesystem.GetGlobalFileSystem()),
	)
	require.NoError(t, err)
}

func TestValidateYAMLFileAgainstSchemaFS(t *testing.T) {
	err := ValidateYAMLFileAgainstSchemaFS(context.Background(), newEmbeddedFilesystem(t), path.Join("testdata", "valid.yaml"), nil, *newValidSchema(t, newEmbeddedFilesystem(t)))
	require.NoError(t, err)
}

func TestValidateYAMLFileAgainstSchemaFSWithOptions(t *testing.T) {
	err := ValidateYAMLFileAgainstSchemaFSWithOptions(
		context.Background(),
		newEmbeddedFilesystem(t),
		path.Join("testdata", "valid.yaml"),
		nil,
		WithTitle("person"),
		WithLocalPath(path.Join("testdata", "person.schema.json")),
		WithFilesystem(newEmbeddedFilesystem(t)),
	)
	require.NoError(t, err)
}

func TestValidateJSONFileAgainstSchema_InvalidJSONFile(t *testing.T) {
	err := ValidateJSONFileAgainstSchema(context.Background(), path.Join("testdata", "invalid.json"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrMarshalling)
}

func TestValidateJSONFileAgainstSchema_InvalidDataFile(t *testing.T) {
	err := ValidateJSONFileAgainstSchema(context.Background(), path.Join("testdata", "does_not_match.json"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestValidateYAMLFileAgainstSchema_InvalidDataFile(t *testing.T) {
	err := ValidateYAMLFileAgainstSchema(context.Background(), path.Join("testdata", "does_not_match.yaml"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}
