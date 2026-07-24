package annotations

import (
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// AnnotationLogger formats annotations and emits them through an underlying
// logger.
//
// The logger delegates final emission to an existing [logs.Loggers]
// implementation and is therefore suitable for any sink already supported by
// the logs package.
type AnnotationLogger struct {
	baselogs.Loggers
	formatter IFormatter
}

// NewLogger creates an annotation logger backed by baseLogger.
func NewLogger(baseLogger baselogs.Loggers, formatter IFormatter) (*AnnotationLogger, error) {
	logger := &AnnotationLogger{Loggers: baseLogger, formatter: formatter}
	if err := logger.Check(); err != nil {
		return nil, err
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
	switch annotation.Severity {
	case SeverityError:
		l.LogError(line)
	default:
		l.Log(line)
	}
	return nil
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
