package annotations

import (
	"fmt"
	"io"
	"os"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// AnnotationLogger formats annotations and emits them through an underlying
// logger.
//
// Annotations are for human usage rather than machine usage like other logs.
// Prefer structured logging over annotations if expecting the result to be parsed by machines only
//
// The logger delegates final emission to an existing [logs.Loggers]
// implementation and is therefore suitable for any sink already supported by
// the logs package that provides access to the underlying writers.
type AnnotationLogger struct {
	annotationWriter io.Writer
	baselogs.Loggers
	formatter IFormatter
}

// NewLogger creates an annotation logger backed by baseLogger.
func NewLogger(baseLogger baselogs.Loggers, formatter IFormatter) (*AnnotationLogger, error) {
	logger := &AnnotationLogger{Loggers: baseLogger, formatter: formatter, annotationWriter: os.Stdout}
	if err := logger.Check(); err != nil {
		return nil, err
	}

	if genericLogger, ok := baseLogger.(baselogs.ILoggerWithUnderlyingWriters); ok {
		logger.annotationWriter, _ = genericLogger.Writers()
	} else {
		baseLogger.Log("the chosen logger does not provide support for annotations, stdout will be used for annotations")
	}

	return logger, nil
}

func (l *AnnotationLogger) Check() error {
	if reflection.IsEmpty(l.Loggers) {
		return commonerrors.ErrNoLogger
	}
	if l.formatter == nil {
		return commonerrors.UndefinedVariable("annotation formatter")
	}
	return l.Loggers.Check()
}

// WriteAnnotation writes annotation using the configured formatter.
func (l *AnnotationLogger) WriteAnnotation(annotation *Annotation) error {
	if err := l.Check(); err != nil {
		return err
	}
	if reflection.IsEmpty(annotation) {
		return commonerrors.UndefinedVariable("annotation")
	}
	line := l.formatter.Format(annotation)
	_, err := fmt.Fprintln(l.annotationWriter, line)
	return err
}

// WriteError writes an error-level annotation.
func (l *AnnotationLogger) WriteError(message string, options ...AnnotationOption) error {
	annotation := newAnnotation(SeverityError, message, options...)
	return l.WriteAnnotation(&annotation)
}

// WriteWarning writes a warning-level annotation.
func (l *AnnotationLogger) WriteWarning(message string, options ...AnnotationOption) error {
	annotation := newAnnotation(SeverityWarning, message, options...)
	return l.WriteAnnotation(&annotation)
}

// WriteNotice writes a notice-level annotation.
func (l *AnnotationLogger) WriteNotice(message string, options ...AnnotationOption) error {
	annotation := newAnnotation(SeverityNotice, message, options...)
	return l.WriteAnnotation(&annotation)
}

var _ IAnnotationLogger = (*AnnotationLogger)(nil)
