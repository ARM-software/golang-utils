// Package pipeline provides sequential, typed processing pipelines.
//
// A pipeline is an ordered chain of steps where the output of one step becomes
// the input of the next step. Pipelines are useful when a workflow has a clear
// data flow and each stage builds on the result of the previous one.
//
// In practice, a pipeline is a good fit for workflows such as:
//   - parse -> validate -> enrich -> render
//   - load config -> normalise -> apply defaults -> validate
//   - read request -> transform state -> persist state
//
// This package exposes two related models:
//   - [SimplePipeline] for same-type flows such as `T -> T -> T`
//   - [Pipeline] for heterogeneous flows such as `string -> Model -> []byte`
//
// Pipelines differ from execution groups and transform groups in intent:
//
//   - A pipeline models a sequence of dependent stages. Step order matters
//     because each step consumes the result produced by the previous step.
//   - An execution group models a collection of functions to run according to
//     group execution rules such as sequential execution, concurrency, stop on
//      the first error, or error joining. Functions in an execution group are not
//     inherently part of a single data flow.
//   - A transform group applies the same transform function to one or more
//     independent inputs and collects the produced outputs. It is useful when
//     each input can be transformed in isolation.
//
// A simple rule of thumb is:
//   - use a pipeline when step N needs the value produced by step N-1
//   - use an execution group when you want to coordinate function execution
//   - use a transform group when you want to apply one transform to many items
//
// The two concepts can also be combined. For example, a pipeline step may use a
// transform group internally to fan out work for a single stage, while the
// outer pipeline still preserves the higher-level sequential workflow.
package pipeline
