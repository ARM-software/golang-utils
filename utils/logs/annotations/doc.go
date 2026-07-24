// Package annotations provides loggers for emitting structured build and CI
// annotations such as errors, warnings, and notices.
//
// Unlike ordinary log lines, annotations are intended for systems that can
// highlight issues in a richer way, for example by attaching messages to files
// and line numbers or surfacing build problems prominently in the UI.
//
// The package provides a generic annotation model plus platform-specific
// formatters for common CI systems. Annotation loggers are built on top of the
// repository's [logs.Loggers] abstraction so they can reuse existing logging
// sinks.
//
// References:
//   - GitHub Actions workflow commands:
//     https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands
//   - Azure DevOps logging commands:
//     https://learn.microsoft.com/en-us/azure/devops/pipelines/scripts/logging-commands
//   - TeamCity service messages:
//     https://www.jetbrains.com/help/teamcity/service-messages.html
package annotations
