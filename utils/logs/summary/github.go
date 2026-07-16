package summary

import "github.com/ARM-software/golang-utils/utils/environment"

const (
	// GitHubStepSummaryEnvironmentVariable is the environment variable pointing to
	// the current GitHub Actions step summary file.
	GitHubStepSummaryEnvironmentVariable = "GITHUB_STEP_SUMMARY"
)

// NewGitHubSummaryLogger creates a summary logger backed by the GitHub Actions
// step summary file. GitHub renders the resulting file as Markdown.
//
// Reference:
//   - https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-job-summary
func NewGitHubSummaryLogger(path string, loggerSource string) (logger *FileSummaryLogger, err error) {
	return NewFileSummaryLogger(path, loggerSource)
}

// NewGitHubSummaryLoggerFromEnvironment creates a GitHub-backed summary logger
// from `GITHUB_STEP_SUMMARY`.
func NewGitHubSummaryLoggerFromEnvironment(loggerSource string) (logger *FileSummaryLogger, err error) {
	envvar, err := environment.NewCurrentEnvironment().GetEnvironmentVariable(GitHubStepSummaryEnvironmentVariable)
	if err != nil {
		return
	}
	return NewGitHubSummaryLogger(envvar.GetValue(), loggerSource)
}
