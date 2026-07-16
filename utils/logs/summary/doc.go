// Package summary provides summary-oriented loggers and writers.
//
// Unlike regular streaming loggers, summary loggers target destinations that
// accumulate a rendered summary. This is useful when a concise human-readable
// report should be presented separately from raw logs, for example in CI
// systems, automated reports, or workflow dashboards.
//
// Summary outputs are typically append-oriented and support a richer format than
// plain line logs. This makes them suitable for status sections, bullet lists,
// result tables, and other compact outputs that are easier to scan than a full
// log stream. The package APIs write plain strings; destinations may interpret
// those strings as Markdown or another rendered format.
//
// The package provides two concrete implementations:
//   - an in-memory summary logger backed by the repository's plain string logger
//   - a file-backed summary logger that flushes accumulated summary content to a file
//
// References:
//   - CommonMark specification:
//     https://spec.commonmark.org/
//   - Markdown guide:
//     https://www.markdownguide.org/basic-syntax/
package summary
