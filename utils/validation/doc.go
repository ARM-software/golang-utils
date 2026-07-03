// Package validation provides validation helpers built on top of
// ozzo-validation.
//
// The package contains two complementary layers:
//   - general-purpose Go validation helpers that extend ozzo-validation and its
//     `is` package
//   - schema-oriented helpers whose naming and behaviour align with external
//     schema ecosystems where that maps cleanly to runtime validation rules
//
// In particular, the schema-oriented helpers aim to cover practical rule
// vocabulary commonly found in:
//   - JSON Schema
//   - OpenAPI / Kubernetes CRD schema extensions
//   - Protovalidate and protobuf validation rule sets
//   - XSD-style structural constraints
//   - Avro- and Thrift-style data shape constraints where they make sense in a
//     runtime Go validation package
//
// The package does not try to replace full schema compilers. Instead, it
// provides reusable validation rules that are convenient when callers already
// have decoded Go values and want rule-level validation with ozzo semantics.
//
// Upstream projects:
//   - ozzo-validation: https://github.com/go-ozzo/ozzo-validation
//   - ozzo-validation/is: https://github.com/go-ozzo/ozzo-validation/tree/master/is
//
// Documentation:
//   - https://go-ozzo.github.io/ozzo-validation/
package validation
