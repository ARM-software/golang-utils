package saga

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func newTestStep(t *testing.T, ctlr *gomock.Controller, identifier IActionIdentifier, expectExecution, expectCompensation, failExecution bool) ITransactionStep {
	t.Helper()
	step := NewMockITransactionStep(ctlr)
	step.EXPECT().GetID().Return(identifier).AnyTimes()
	if expectCompensation {
		step.EXPECT().Compensate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ any) error {
			fmt.Println("compensating ", identifier)
			return nil
		}).Times(1)
	}
	if expectExecution {
		if failExecution {
			step.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ any) error {
				fmt.Println("executing ", identifier)
				return commonerrors.New(commonerrors.ErrUnexpected, "failed execution")
			}).MinTimes(1)
		} else {
			step.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ any) error {
				fmt.Println("executing ", identifier)
				return nil
			}).MinTimes(1)
		}
	}
	return step
}

func TestSaga_AllStepsSucceed_NoCompensation(t *testing.T) {
	ctx := context.Background()
	orch := NewMinimalSaga(NoStepArguments())
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	step1 := newTestStep(t, ctlr, NewStepIdentifier("step1", faker.DomainName()), true, false, false)
	step2 := newTestStep(t, ctlr, NewStepIdentifier("step2", faker.DomainName()), true, false, false)
	step3 := newTestStep(t, ctlr, NewStepIdentifier("step3", faker.DomainName()), true, false, false)

	orch.RegisterFunction(step1, step2, step3)

	assert.Equal(t, 3, orch.Len())

	err := orch.Execute(ctx)
	require.NoError(t, err)
}

func TestSaga_AllStepsSucceed_NewFrom(t *testing.T) {
	ctx := context.Background()
	orch := NewMinimalSaga(NoStepArguments())
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	step1 := newTestStep(t, ctlr, NewStepIdentifier("step1", faker.DomainName()), true, false, false)
	step2 := newTestStep(t, ctlr, NewStepIdentifier("step2", faker.DomainName()), true, false, false)
	step3 := newTestStep(t, ctlr, NewStepIdentifier("step3", faker.DomainName()), true, false, false)

	orch.RegisterFunction(step1, step2, step3)

	assert.Equal(t, 3, orch.Len())
	norch := orch.NewSaga(NoStepArguments())
	assert.Equal(t, 3, norch.Len())
	err := orch.Execute(ctx)
	require.NoError(t, err)

	err = norch.Execute(ctx)
	require.NoError(t, err)
	assert.NotEqual(t, orch.getArgs().GetIdemKey(), norch.getArgs().GetIdemKey())
}

func TestSaga_AllStepsSucceed_Clone(t *testing.T) {
	ctx := context.Background()
	orch := NewMinimalSaga(NoStepArguments())
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	step1 := newTestStep(t, ctlr, NewStepIdentifier("step1", faker.DomainName()), true, false, false)
	step2 := newTestStep(t, ctlr, NewStepIdentifier("step2", faker.DomainName()), true, false, false)
	step3 := newTestStep(t, ctlr, NewStepIdentifier("step3", faker.DomainName()), true, false, false)

	orch.RegisterFunction(step1, step2, step3)

	assert.Equal(t, 3, orch.Len())
	corch := orch.Clone()
	assert.Equal(t, 3, corch.Len())

	err := corch.Execute(ctx)
	require.NoError(t, err)
}

func TestSaga_FailureTriggersCompensation(t *testing.T) {
	t.Run("partial process", func(t *testing.T) {
		ctx := context.Background()
		orch := NewMinimalSaga(NoStepArguments())
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		step1 := newTestStep(t, ctlr, NewStepIdentifier("step1", faker.DomainName()), true, true, false)
		step2 := newTestStep(t, ctlr, NewStepIdentifier("step2", faker.DomainName()), true, true, true)
		step3 := newTestStep(t, ctlr, NewStepIdentifier("step3", faker.DomainName()), false, false, false)

		orch.RegisterFunction(step1, step2, step3)

		err := orch.Execute(ctx)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
	})
	t.Run("full process", func(t *testing.T) {
		ctx := context.Background()
		orch := NewMinimalSaga(NoStepArguments())
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		step1 := newTestStep(t, ctlr, NewStepIdentifier("step1", faker.DomainName()), true, true, false)
		step2 := newTestStep(t, ctlr, NewStepIdentifier("step2", faker.DomainName()), true, true, false)
		step3 := newTestStep(t, ctlr, NewStepIdentifier("step3", faker.DomainName()), true, true, true)

		orch.RegisterFunctions(slices.Values([]ITransactionStep{step1, step2, step3}))

		err := orch.Execute(ctx)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
	})
}

func TestSaga_Empty_NoError(t *testing.T) {
	ctx := context.Background()
	orch := NewMinimalSaga(NoStepArguments())

	assert.Zero(t, orch.Len())

	require.NoError(t, orch.Execute(ctx))
}
