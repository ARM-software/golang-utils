// Deprecated: use github.com/ARM-software/golang-utils/utils/flow/transaction/saga instead.
//
// Package saga provides an implementation for the [SAGA pattern](https://microservices.io/patterns/data/saga.html) for [transaction like distribution processes](https://learn.microsoft.com/en-us/azure/architecture/patterns/saga) across multiple services without relying on a global ACID transaction.
// The SAGA orchestration pattern breaks a distributed business operation into a sequence of local transactions coordinated by a central orchestrator. Each step executes independently, and if any step fails, the orchestrator triggers compensating transactions in reverse order to undo completed work, following the [Compensating Transaction pattern](https://en.wikipedia.org/wiki/Compensating_transaction?utm_source=chatgpt.com) (https://learn.microsoft.com/en-us/azure/architecture/patterns/compensating-transaction).
// To make sagas safe across retries, failures, or orchestrator crashes, compensating transactions must themselves be designed to be idempotent and retryable. That way if a compensation fails (or the orchestrator crashes mid-rollback), they can be resumed/ retried without risk of double-undo or inconsistent state.  For that purpose, idempotency keys are used, ensuring repeated “execute” or “compensate” calls do not duplicate effects, as described in [idempotency practices](https://brandur.org/idempotency-keys)
package saga

import newsaga "github.com/ARM-software/golang-utils/utils/flow/transaction/saga"

// IActionArguments is kept for backward compatibility.
type IActionArguments = newsaga.IActionArguments

// IActionIdentifier is kept for backward compatibility.
type IActionIdentifier = newsaga.IActionIdentifier

// ITransactionStep is kept for backward compatibility.
type ITransactionStep = newsaga.ITransactionStep

// ISagaOrchestrator is kept for backward compatibility.
type ISagaOrchestrator = newsaga.ISagaOrchestrator

// MinimalSaga is kept for backward compatibility.
type MinimalSaga = newsaga.MinimalSaga

func NewMinimalSaga(args IActionArguments) *MinimalSaga {
	return newsaga.NewMinimalSaga(args)
}

func NewStepIdentifier(name, namespace string) IActionIdentifier {
	return newsaga.NewStepIdentifier(name, namespace)
}

func NewStepArgumentsWithIdempotentKey(idemKey string, args map[string]any) IActionArguments {
	return newsaga.NewStepArgumentsWithIdempotentKey(idemKey, args)
}

func NewStepArguments(args map[string]any) IActionArguments {
	return newsaga.NewStepArguments(args)
}

func NoStepArguments() IActionArguments {
	return newsaga.NoStepArguments()
}
