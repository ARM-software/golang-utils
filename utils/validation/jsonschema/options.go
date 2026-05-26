package jsonschema

import (
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// SchemaOption configures a Schema during construction.
type SchemaOption func(*Schema) *Schema

// DefaultSchema returns a Schema initialised with the package defaults.
//
// The default filesystem is the global filesystem, and the schema ID is derived
// from the local path when it has not been set explicitly.
func DefaultSchema() *Schema {
	return (&Schema{}).Default()
}

// Default applies the package defaults to a Schema.
func (s *Schema) Default() *Schema {
	if s == nil {
		s = &Schema{}
	}
	if s.Filesystem == nil {
		s.Filesystem = filesystem.GetGlobalFileSystem()
	}
	if s.Limits == nil {
		s.Limits = filesystem.NoLimits()
	}
	s.ID = schemaID(s.ID, s.LocalPath)
	return s
}

// Apply applies a single option to a Schema and then reapplies defaults.
func (s *Schema) Apply(option SchemaOption) *Schema {
	if option == nil {
		return s.Default()
	}
	return option(s).Default()
}

// NewJSONSchemaFile constructs a Schema from the supplied options.
//
// Options are applied in order, after which the package defaults are
// materialised.
func NewJSONSchemaFile(options ...SchemaOption) *Schema {
	schema := DefaultSchema()
	collection.ForEach(options, func(option SchemaOption) {
		schema.Apply(option)
	})
	return schema.Default()
}

// WithTitle sets the human-readable title of a Schema.
func WithTitle(title string) SchemaOption {
	return func(schema *Schema) *Schema {
		if schema == nil {
			schema = DefaultSchema()
		}
		schema.Title = strings.TrimSpace(title)
		return schema
	}
}

// WithLocalPath sets the local path of a Schema file.
//
// The path is trimmed, and the schema ID is recalculated with schemaID when no
// explicit ID has been provided.
func WithLocalPath(localPath string) SchemaOption {
	return func(schema *Schema) *Schema {
		if schema == nil {
			schema = DefaultSchema()
		}
		schema.LocalPath = localPath
		schema.ID = schemaID(schema.ID, schema.LocalPath)
		return schema
	}
}

// WithID sets the schema ID of a Schema.
//
// If the supplied ID is empty, schemaID falls back to the current local path.
func WithID(id string) SchemaOption {
	return func(schema *Schema) *Schema {
		if schema == nil {
			schema = DefaultSchema()
		}
		schema.ID = schemaID(id, schema.LocalPath)
		return schema
	}
}

// WithFilesystem sets the filesystem used to load the Schema file.
//
// A nil filesystem falls back to the global filesystem, and the current local
// path is cleaned for that filesystem.
func WithFilesystem(fs filesystem.FS) SchemaOption {
	return func(schema *Schema) *Schema {
		if schema == nil {
			schema = DefaultSchema()
		}
		if fs == nil {
			fs = filesystem.GetGlobalFileSystem()
		}
		schema.Filesystem = fs
		schema.LocalPath = filesystem.FilePathClean(schema.Filesystem, schema.LocalPath)
		return schema
	}
}

// WithFileLimits sets the filesystem read limits used when loading the Schema
// file.
//
// A nil limits value falls back to filesystem.NoLimits().
func WithFileLimits(limits filesystem.ILimits) SchemaOption {
	return func(schema *Schema) *Schema {
		if schema == nil {
			schema = DefaultSchema()
		}
		if limits == nil {
			limits = filesystem.NoLimits()
		}
		schema.Limits = limits
		return schema
	}
}
