// Package flow provides a home for flow-oriented orchestration primitives.
//
// The package is intentionally broader than a single execution model. It is
// intended to encompass linear chains, pipelines, DAGs, state machines, and
// workflows over time without committing the module structure to any one of
// those models upfront.
//
// Current subpackages include:
//   - [github.com/ARM-software/golang-utils/utils/flow/step] for reusable flow steps
//   - [github.com/ARM-software/golang-utils/utils/flow/pipeline] for linear typed pipelines
//   - [github.com/ARM-software/golang-utils/utils/flow/transaction/saga] for saga-style transaction orchestration
package flow
