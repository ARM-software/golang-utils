// Package casing rewrites identifier-like strings according to declarative
// token replacement rules.
//
// "Identifier-like strings" here means names that are structured as code or
// configuration identifiers rather than arbitrary prose, for example:
//   - camelCase names such as `apiClient` or `sourceName`
//   - PascalCase names such as `ApiClient` or `SourceName`
//   - mixed acronym-containing identifiers such as `HttpAPIClient`
//
// It is intended for cases where callers need to preserve the overall casing
// shape of such identifiers while forcing specific words to use a canonical
// replacement and exempting other whole words through an exception list.
//
// This differs from simpler alternatives such as:
//   - `strings.Replacer`, which is useful for literal substring replacement but
//     does not understand identifier word boundaries or case reconstruction
//   - regexp-based rewriting, which can express some token patterns but becomes
//     awkward when replacement depends on identifier-style word splitting,
//     acronym policies, and explicit exception lists
//
// The package therefore combines identifier splitting with rule-based word
// replacement, rather than treating the input as an arbitrary free-form string.
//
// Unlike libraries that automatically apply a built-in Go initialism list, this
// package only applies that behaviour when explicitly requested. Callers can opt
// into it by creating a replacer from [InitialismRules] or by reusing the
// ready-made [InitialismReplacer]. This keeps the package predictable for
// domain-specific casing policies while still providing built-in optional Go
// initialism support when wanted.
//
// Typical uses include:
//   - normalising known acronyms in generated identifiers
//   - preserving exception words that should not be rewritten
//   - applying a shared replacement policy across tools or generators
//
// Related references:
//   - Go initialisms guidance in effective naming:
//     https://go.dev/wiki/CodeReviewComments#initialisms
//   - `ettle/strcase`, which includes a broad casing test corpus and built-in
//     Go-initialism-aware conversions:
//     https://github.com/ettle/strcase
//
// The case-conversion helpers in this package accept `...*Replacer` so callers
// can use one common helper shape whether they want replacement or not. Only
// the first replacer is used; passing more than one replacer is not supported.
package casing
