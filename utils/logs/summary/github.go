package summary

import "github.com/ARM-software/golang-utils/utils/environment"

const (
	// GitHubStepSummaryEnvironmentVariable is the environment variable pointing to
	// the current GitHub Actions step summary file.
	GitHubStepSummaryEnvironmentVariable = "GITHUB_STEP_SUMMARY"
)

// NewGitHubSummaryLogger creates a summary logger backed by the file path stored
// in `GITHUB_STEP_SUMMARY`.
//
// GitHub renders that file as Markdown on the workflow summary page.
//
// Reference:
//   - https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-job-summary
func NewGitHubSummaryLogger(loggerSource string) (logger *FileSummaryLogger, err error) {
	envvar, err := environment.NewCurrentEnvironment().GetEnvironmentVariable(GitHubStepSummaryEnvironmentVariable)
	if err != nil {
		return
	}
	return NewFileSummaryLogger(envvar.GetValue(), loggerSource)
}
