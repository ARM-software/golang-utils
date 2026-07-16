// Package summary provides loggers for producing concise human-readable
// summaries rather than full diagnostic log streams.
//
// Unlike regular streaming loggers, summary loggers target destinations that
// accumulate a rendered summary. This is useful when a concise human-readable
// report should be presented separately from raw logs, for example in CI
// systems, automated reports, or workflow dashboards.
//
// The package writes plain strings, so the destination can render the content
// directly or interpret it using Markdown. This is helpful for status sections,
// bullet lists, check results, and short report content that should stay easy
// to scan.
//
// References:
//   - CommonMark specification: https://spec.commonmark.org/
//   - Markdown guide: https://www.markdownguide.org/basic-syntax/
package summary
