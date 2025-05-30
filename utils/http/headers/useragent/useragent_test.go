package useragent

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestAddValuesToUserAgent(t *testing.T) {
	userAgent, err := GenerateUserAgentValue(faker.Name(), faker.Word(), faker.Sentence())
	require.NoError(t, err)
	newAgent := AddValuesToUserAgent(userAgent, faker.Word())
	assert.Contains(t, newAgent, userAgent)
	elem := faker.Word()
	newAgent = AddValuesToUserAgent("   ", elem)
	assert.Equal(t, elem, newAgent)
}

func TestGenerateUserAgentValue(t *testing.T) {
	userAgent, err := GenerateUserAgentValue("", "", "")
	errortest.AssertError(t, err, commonerrors.ErrUndefined, commonerrors.ErrInvalid)
	assert.Empty(t, userAgent)
	userAgent, err = GenerateUserAgentValue("      ", "          ", faker.Sentence())
	errortest.AssertError(t, err, commonerrors.ErrUndefined, commonerrors.ErrInvalid)
	assert.Empty(t, userAgent)
	userAgent, err = GenerateUserAgentValue(faker.Name(), "", faker.Sentence())
	errortest.AssertError(t, err, commonerrors.ErrUndefined, commonerrors.ErrInvalid)
	assert.Empty(t, userAgent)
	userAgent, err = GenerateUserAgentValue(faker.Name(), faker.Word(), faker.Sentence())
	assert.NoError(t, err)
	assert.NotEmpty(t, userAgent)
}
