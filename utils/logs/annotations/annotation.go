package annotations

import (
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
)

// Severity identifies the severity of an annotation.
type Severity int

//go:generate go tool enumer -type=Severity -text -json -yaml

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityNotice
)

// Annotation describes a structured annotation message.
type Annotation struct {
	// Severity identifies whether the annotation is an error, warning, or notice.
	Severity Severity
	// Message is the human-readable annotation text.
	Message string
	// File is the file path associated with the annotation, if any.
	File string
	// Line is the 1-based line number associated with the annotation, if any.
	Line *int
	// Column is the 1-based column number associated with the annotation, if any.
	Column *int
}

// AnnotationOption configures an annotation.
type AnnotationOption func(*Annotation) *Annotation

// WithFile sets the file path associated with an annotation.
func WithFile(path string) AnnotationOption {
	return func(annotation *Annotation) *Annotation {
		annotation.File = path
		return annotation
	}
}

// WithLine sets the line associated with an annotation.
func WithLine(line int) AnnotationOption {
	return func(annotation *Annotation) *Annotation {
		annotation.Line = field.ToOptionalInt(line)
		return annotation
	}
}

// WithColumn sets the column associated with an annotation.
func WithColumn(column int) AnnotationOption {
	return func(annotation *Annotation) *Annotation {
		annotation.Column = field.ToOptionalInt(column)
		return annotation
	}
}

func newAnnotation(severity Severity, message string, options ...AnnotationOption) Annotation {
	annotation := Annotation{Severity: severity, Message: message}
	collection.ForEach(options, func(option AnnotationOption) {
		if option != nil {
			annotation = *option(&annotation)
		}
	})
	return annotation
}
