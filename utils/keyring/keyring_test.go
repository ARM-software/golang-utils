package keyring

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

type TestCfg struct {
	Test1 string
	Test2 int `mapstructure:"test64"`
	Test3 uint
	Test4 float64
	Test5 bool
	Test6 time.Duration
	Test7 time.Time
	Test8 time.Location
}

type TestCfg1 struct {
	Test1 string
	Test2 time.Duration
	Test3 TestCfg `mapstructure:"subtest_test"`
}

func TestKeyring(t *testing.T) {
	expected := TestCfg1{}
	require.NoError(t, faker.FakeData(&expected))
	prefix := faker.Word()
	err := Store[TestCfg1](context.Background(), prefix, &expected)
	errortest.AssertError(t, err, nil, commonerrors.ErrUnsupported)
	if commonerrors.Any(err, commonerrors.ErrUnsupported) {
		t.Skip("keyring is not supported")
	}
	actual := TestCfg1{}
	require.NoError(t, Fetch[TestCfg1](context.Background(), prefix, &actual))
	assert.EqualExportedValues(t, expected, actual)
	require.NoError(t, Clear(context.Background(), prefix))
	require.NoError(t, Fetch[TestCfg1](context.Background(), prefix, &actual))
	assert.EqualExportedValues(t, expected, actual)
	actual2 := TestCfg1{}
	require.NoError(t, Fetch[TestCfg1](context.Background(), prefix, &actual2))
	assert.Empty(t, actual2)
	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, Fetch[TestCfg1](ctx, prefix, &actual), commonerrors.ErrCancelled)
		errortest.AssertError(t, Store[TestCfg1](ctx, prefix, &actual), commonerrors.ErrCancelled)
	})
}
