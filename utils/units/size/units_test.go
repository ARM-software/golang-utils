package size

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSizes(t *testing.T) {
	sizes := []float64{B, KB, MB, GB, TB, PB, EB, ZB, YB}
	for i := range sizes {
		if i > 0 {
			assert.Equal(t, sizes[i], 1000*sizes[i-1])
		}
	}
	sizes = []float64{B, KiB, MiB, GiB, TiB, PiB, EiB, ZiB, YiB}
	for i := range sizes {
		if i > 0 {
			assert.Equal(t, sizes[i], 1024*sizes[i-1])
		}
	}
	assert.Equal(t, KiB, float64(1024))
	assert.Equal(t, KiB, float64(1<<10))
	assert.Equal(t, GiB, float64(1<<30))
	assert.Equal(t, 10*GiB, float64(10<<30))
	assert.Equal(t, MiB, float64(1<<20))
}
