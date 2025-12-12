package saga

import (
	"context"
	"iter"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

var _ ISagaOrchestrator = &MinimalSaga{}

// NewMinimalSaga
func NewMinimalSaga(args IActionArguments) *MinimalSaga {
	s := &MinimalSaga{args: args}
	s.defineCompensationStore()
	s.defineTransactionStore()
	return s
}

type MinimalSaga struct {
	args         IActionArguments
	transaction  parallelisation.IExecutionGroup[ITransactionStep]
	compensation parallelisation.IExecutionGroup[ITransactionStep]
}

func (s *MinimalSaga) RegisterFunction(function ...ITransactionStep) {
	s.GetTransaction().RegisterFunction(function...)
}

func (s *MinimalSaga) RegisterFunctions(function iter.Seq[ITransactionStep]) {
	s.GetTransaction().RegisterFunctions(function)
}

func (s *MinimalSaga) CopyFunctions(g parallelisation.IExecutionGroup[ITransactionStep]) {
	if g == nil {
		return
	}
	s.GetTransaction().CopyFunctions(g)
}

func (s *MinimalSaga) defineCompensationStore() {
	s.compensation = parallelisation.NewExecutionGroup[ITransactionStep](func(ctx context.Context, element ITransactionStep) error {
		if element == nil {
			return commonerrors.UndefinedVariable("transaction step")
		}
		return element.Compensate(ctx, s.getArgs())
	}, parallelisation.JoinErrors, parallelisation.SequentialInReverse, parallelisation.StopOnFirstError, parallelisation.OnlyOnce)
}

func (s *MinimalSaga) GetCompensation() parallelisation.IExecutionGroup[ITransactionStep] {
	return s.compensation
}
func (s *MinimalSaga) GetTransaction() parallelisation.IExecutionGroup[ITransactionStep] {
	return s.transaction
}

func (s *MinimalSaga) getArgs() IActionArguments {
	return s.args
}

func (s *MinimalSaga) defineTransactionStore() {
	s.transaction = parallelisation.NewExecutionGroup[ITransactionStep](func(ctx context.Context, element ITransactionStep) error {
		if element == nil {
			return commonerrors.UndefinedVariable("transaction step")
		}
		err := element.Execute(ctx, s.getArgs())
		s.GetCompensation().RegisterFunction(element)
		return err
	}, parallelisation.JoinErrors, parallelisation.Sequential, parallelisation.StopOnFirstError, parallelisation.RetainAfterExecution)
}

// Clone returns a clone of the orchestrator and its execution state.
func (s *MinimalSaga) Clone() parallelisation.IExecutionGroup[ITransactionStep] {
	clone := &MinimalSaga{
		args:         s.getArgs(),
		transaction:  s.transaction.Clone(),
		compensation: s.compensation.Clone(),
	}
	return clone
}

func (s *MinimalSaga) NewSaga(args IActionArguments) *MinimalSaga {
	newSaga := &MinimalSaga{
		args:        args,
		transaction: s.transaction.Clone(),
	}
	newSaga.defineCompensationStore()
	return newSaga
}

func (s *MinimalSaga) Execute(ctx context.Context) error {
	err := s.GetTransaction().Execute(ctx)
	if err == nil {
		return nil
	}
	return commonerrors.Join(err, s.GetCompensation().Execute(ctx))
}

func (s *MinimalSaga) Len() int {
	return s.GetTransaction().Len()
}
