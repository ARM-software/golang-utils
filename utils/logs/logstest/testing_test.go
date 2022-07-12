package logstest

import (
	"testing"

	"github.com/bxcodec/faker/v3"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestNewNullTestLogger(t *testing.T) {
	logger := NewNullTestLogger()
	logger.WithValues("foo", "bar").Info(faker.Sentence())
	logger.Error(commonerrors.ErrUnexpected, faker.Sentence(), faker.Word(), faker.Name())
}

func TestNewTestLogger(t *testing.T) {
	logger := NewTestLogger(t)
	logger.Info(faker.Sentence())
	logger.Info(faker.Sentence(), "foo", "bar")
	logger.Error(commonerrors.ErrUnexpected, faker.Sentence(), faker.Word(), faker.Name())
}

func TestNewStdTestLogger(t *testing.T) {
	logger := NewStdTestLogger()
	logger.WithValues("foo", "bar").Info(faker.Sentence())
	logger.Error(commonerrors.ErrUnexpected, faker.Sentence(), faker.Word(), faker.Name())
}
