package multiplication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumbers(t *testing.T) {
	assert.Equal(t, Kilo, float64(1000))
	assert.Equal(t, Mega, float64(1000000))

}
