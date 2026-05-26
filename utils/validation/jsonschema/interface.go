package jsonschema

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

//go:generate go tool mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/validation/$GOPACKAGE ISchemaValidator

// ISchemaValidator validates content against a compiled JSON
// Schema definition.
//
// Implementations may support different source formats such as JSON or YAML,
// but all expose the same three entry points: validating already-loaded
// content, validating a file from the global filesystem, and validating a file
// from a supplied filesystem.
type ISchemaValidator interface {
	// ValidateContent validates already-loaded content against the compiled
	// schema definition.
	ValidateContent(context.Context, any) error
	// ValidateFile validates a file from the global filesystem against the
	// compiled schema definition.
	ValidateFile(ctx context.Context, filepath string) error
	// ValidateFileWithLimits validates a file from the global filesystem against the
	// compiled schema definition.
	ValidateFileWithLimits(ctx context.Context, filepath string, fileLimits filesystem.ILimits) error
	// ValidateFileInFS validates a file from the supplied filesystem against the
	// compiled schema definition.
	ValidateFileInFS(ctx context.Context, fs filesystem.FS, filepath string) error
	// ValidateFileInFSWithLimits validates a file from the supplied filesystem against the
	// compiled schema definition.
	ValidateFileInFSWithLimits(ctx context.Context, fs filesystem.FS, filepath string, fileLimits filesystem.ILimits) error
}
