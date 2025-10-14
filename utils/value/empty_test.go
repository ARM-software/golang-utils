package value

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	type testInterface interface {
	}
	var testEmptyPtr testInterface
	emptyStr := ""
	whiteSpace := "                  "

	aFilledChannel := make(chan struct{}, 1)
	aFilledChannel <- struct{}{}
	tests := []struct {
		value                  interface{}
		isEmpty                bool
		differsFromAssertEmpty bool
	}{
		{
			value:   nil,
			isEmpty: true,
		},
		{
			value:   0,
			isEmpty: true,
		},
		{
			value:   uint(0),
			isEmpty: true,
		},
		{
			value:   float64(0),
			isEmpty: true,
		},
		{
			value:   "",
			isEmpty: true,
		},
		{
			value:                  "                                   ",
			isEmpty:                true,
			differsFromAssertEmpty: true,
		},
		{
			value:   (*string)(nil),
			isEmpty: true,
		},
		{
			value:   &emptyStr,
			isEmpty: true,
		},
		{
			value:                  &whiteSpace,
			isEmpty:                true,
			differsFromAssertEmpty: true,
		},
		{
			value:   false,
			isEmpty: true,
		},
		{
			value:   []string{},
			isEmpty: true,
		},
		{
			value:   []int64{},
			isEmpty: true,
		},
		{
			value:   []int64{int64(0)},
			isEmpty: false,
		},
		{
			value:   "blah",
			isEmpty: false,
		},
		{
			value:   1,
			isEmpty: false,
		},
		{
			value:   true,
			isEmpty: false,
		},
		{
			value:   testEmptyPtr,
			isEmpty: true,
		},
		{
			value:   map[string]string{},
			isEmpty: true,
		},
		{
			value:   map[string]interface{}{},
			isEmpty: true,
		},
		{
			value:   map[string]interface{}{"foo": "bar"},
			isEmpty: false,
		},
		{
			value:   time.Time{},
			isEmpty: true,
		},
		{
			value:   time.Now(),
			isEmpty: false,
		},
		{
			value:   make(chan struct{}),
			isEmpty: true,
		},
		{
			value:   aFilledChannel,
			isEmpty: false,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("subtest #%v (%v)", i, test.value), func(t *testing.T) {
			assert.Equal(t, test.isEmpty, IsEmpty(test.value))
			if test.isEmpty && !test.differsFromAssertEmpty {
				assert.Empty(t, test.value)
			} else {
				assert.NotEmpty(t, test.value)
			}
		})
	}
}
