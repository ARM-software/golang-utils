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
	return &Schema{
		Title:      "person",
		LocalPath:  path.Join("testdata", "person.schema.json"),
		ID:         "https://example.com/schemas/person.schema.json",
		Filesystem: fs,
	}
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

	err := (&Schema{}).Validate()
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}

func TestSchemaSpecValidate(t *testing.T) {
	require.NoError(t, (&SchemaSpec{ID: "schema.json", Specification: []byte(`{"type":"object"}`)}).Validate())

	err := (&SchemaSpec{}).Validate()
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

func TestValidateRawJSONAgainstSchema(t *testing.T) {
	err := ValidateRawJSONAgainstSchema(context.Background(), map[string]any{"name": "Alice", "count": 2}, nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
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

func TestValidateRawYAMLAgainstSchema(t *testing.T) {
	err := ValidateRawYAMLAgainstSchema(context.Background(), []byte("name: Alice\ncount: 2\n"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
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

func TestValidateJSONFileAgainstSchema(t *testing.T) {
	err := ValidateJSONFileAgainstSchema(context.Background(), path.Join("testdata", "valid.json"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
}

func TestValidateJSONFileAgainstSchemaFS(t *testing.T) {
	err := ValidateJSONFileAgainstSchemaFS(context.Background(), newEmbeddedFilesystem(t), path.Join("testdata", "valid.json"), nil, *newValidSchema(t, newEmbeddedFilesystem(t)))
	require.NoError(t, err)
}

func TestValidateYAMLFileAgainstSchema(t *testing.T) {
	err := ValidateYAMLFileAgainstSchema(context.Background(), path.Join("testdata", "valid.yaml"), nil, *newValidSchema(t, filesystem.GetGlobalFileSystem()))
	require.NoError(t, err)
}

func TestValidateYAMLFileAgainstSchemaFS(t *testing.T) {
	err := ValidateYAMLFileAgainstSchemaFS(context.Background(), newEmbeddedFilesystem(t), path.Join("testdata", "valid.yaml"), nil, *newValidSchema(t, newEmbeddedFilesystem(t)))
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
