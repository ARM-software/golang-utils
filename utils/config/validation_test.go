package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_processMapStructureString(t *testing.T) {
	tests := []struct {
		mapstructureTag      string
		expectedProcessedTag string
	}{
		{},
		{
			mapstructureTag: "         ",
		},
		{
			mapstructureTag: "     -    ",
		},
		{
			mapstructureTag: "    , omitzero      ",
		},
		{
			mapstructureTag: "  ,omitempty  , omitzero    , SQUASH  ",
		},
		{
			mapstructureTag:      "test  ,omitempty  , omitzero    , squash  ",
			expectedProcessedTag: "test",
		},
		{
			mapstructureTag:      "person_name",
			expectedProcessedTag: "person_name",
		},
		{
			mapstructureTag:      "   person_name   ",
			expectedProcessedTag: "person_name",
		},
		{
			mapstructureTag:      "   person_name ,remain  ",
			expectedProcessedTag: "person_name",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.mapstructureTag, func(t *testing.T) {
			assert.Equal(t, test.expectedProcessedTag, processMapStructureString(test.mapstructureTag))
		})
	}
}
