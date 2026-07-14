package summary

import (
	"os"
	"sync"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// GitHubWriter writes summary content to a GitHub Actions step summary file.
type GitHubWriter struct {
	mu   sync.Mutex
	path string
}

// NewGitHubWriter creates a GitHub summary writer for a specific summary file path.
func NewGitHubWriter(path string) (*GitHubWriter, error) {
	writer := &GitHubWriter{path: path}
	if reflection.IsEmpty(path) {
		return nil, commonerrors.New(commonerrors.ErrUndefined, "missing GitHub summary path")
	}
	return writer, nil
}

// NewGitHubWriterFromEnvironment creates a GitHub summary writer from
// `GITHUB_STEP_SUMMARY`.
func NewGitHubWriterFromEnvironment() (*GitHubWriter, error) {
	return NewGitHubWriter(os.Getenv(GitHubStepSummaryEnvironmentVariable))
}

// NewGitHubLogger creates a summary logger backed by a GitHub summary writer.
func NewGitHubLogger(path string, loggerSource string) (Loggers, error) {
	writer, err := NewGitHubWriter(path)
	if err != nil {
		return nil, err
	}
	return NewLogger(writer, loggerSource)
}

// NewGitHubLoggerFromEnvironment creates a summary logger backed by the current
// GitHub Actions step summary file.
func NewGitHubLoggerFromEnvironment(loggerSource string) (Loggers, error) {
	writer, err := NewGitHubWriterFromEnvironment()
	if err != nil {
		return nil, err
	}
	return NewLogger(writer, loggerSource)
}

func (w *GitHubWriter) Close() error {
	return nil
}

func (w *GitHubWriter) WriteMarkdown(markdown string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if reflection.IsEmpty(w.path) {
		return commonerrors.New(commonerrors.ErrUndefined, "missing GitHub summary path")
	}
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	_, err = file.WriteString(markdown)
	return err
}

func (w *GitHubWriter) Clear() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if reflection.IsEmpty(w.path) {
		return commonerrors.New(commonerrors.ErrUndefined, "missing GitHub summary path")
	}
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}
