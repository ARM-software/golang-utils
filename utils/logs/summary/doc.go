// Package summary provides summary-oriented loggers and writers.
//
// Unlike regular streaming loggers, summary loggers target destinations that
// accumulate a rendered summary, typically in Markdown. This is useful for CI
// systems and similar environments where a concise human-readable report should
// be presented separately from raw logs.
//
// One important use case is GitHub Actions job summaries, which are written to
// the file referenced by `GITHUB_STEP_SUMMARY` and rendered as GitHub-flavoured
// Markdown on the workflow run summary page.
//
// References:
//   - GitHub Actions job summaries:
//     https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-job-summary
//   - go-githubactions package summary helpers:
//     https://pkg.go.dev/github.com/sethvargo/go-githubactions
//   - actions-go/toolkit summary builder:
//     https://pkg.go.dev/github.com/actions-go/toolkit/core
package summary
