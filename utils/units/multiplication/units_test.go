package multiplication

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNumbers(t *testing.T) {
	assert.Equal(t, Kilo, float64(1000))
	assert.Equal(t, Mega, float64(1000000))

}
