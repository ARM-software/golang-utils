// Package saga provides an implementation for the [SAGA pattern]((https://microservices.io/patterns/data/saga.html) for [transaction like distribution processes](https://learn.microsoft.com/en-us/azure/architecture/patterns/saga) across multiple services without relying on a global ACID transaction.
// The SAGA orchestration pattern breaks a distributed business operation into a sequence of local transactions coordinated by a central orchestrator. Each step executes independently, and if any step fails, the orchestrator triggers compensating transactions in reverse order to undo completed work, following the [Compensating Transaction pattern](https://en.wikipedia.org/wiki/Compensating_transaction?utm_source=chatgpt.com) (https://learn.microsoft.com/en-us/azure/architecture/patterns/compensating-transaction).
// To make sagas safe across retries, failures, or orchestrator crashes, compensating transactions must themselves be designed to be idempotent and retryable. That way if a compensation fails (or the orchestrator crashes mid-rollback), they can be resumed/ retried without risk of double-undo or inconsistent state.  For that purpose, idempotency keys are used, ensuring repeated “execute” or “compensate” calls do not duplicate effects, as described in [idempotency practices](https://brandur.org/idempotency-keys)
package saga

//go:generate go tool mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/transaction/$GOPACKAGE IActionArguments,IActionIdentifier,ITransactionStep,ISagaOrchestrator
//go:generate go tool mockgen -destination=./mock_test.go -package=saga github.com/ARM-software/golang-utils/utils/transaction/$GOPACKAGE IActionArguments,IActionIdentifier,ITransactionStep,ISagaOrchestrator

import (
	"context"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type IActionArguments interface {
	// GetIdemKey return idempotent key to ensure safe
	// retries and crash recovery, preventing duplicate side effects.
	GetIdemKey() string
	GetArguments() map[string]any
}

type IActionIdentifier interface {
	fmt.Stringer
	GetName() string
	GetNamespace() string
}

// ITransactionStep describes a step in the transaction across a distributed system.
type ITransactionStep interface {
	// GetID returns an identifier of the action
	GetID() IActionIdentifier
	// Execute performs the forward action.
	Execute(ctx context.Context, args IActionArguments) error
	// Compensate performs the compensating/rollback action.
	Compensate(ctx context.Context, args IActionArguments) error
}

// ISagaOrchestrator coordinates a sequence of local transactions (ITransactionStep) to
// achieve an eventually consistent distributed workflow without relying on a
// global ACID transaction. Each step has a forward action (Execute) and a
// compensating action (Compensate). The orchestrator executes steps in order;
// if any step fails, it triggers compensating actions in reverse order to undo
// previously completed steps.
//
// This pattern is useful for long-running or cross-service operations where
// traditional two-phase commit is impractical.
//
// References:
//   - Saga Pattern:
//     https://en.wikipedia.org/wiki/Long-running_transaction
//     https://learn.microsoft.com/en-us/azure/architecture/patterns/saga
//   - Compensating Transaction Pattern:
//     https://learn.microsoft.com/en-us/azure/architecture/patterns/compensating-transaction
//   - Saga Pattern (microservices.io):
//     https://microservices.io/patterns/data/saga.html
type ISagaOrchestrator interface {
	parallelisation.IExecutionGroup[ITransactionStep]
}
