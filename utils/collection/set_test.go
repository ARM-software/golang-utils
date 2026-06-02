package collection

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUniqueBy(t *testing.T) {
	type planet struct {
		Name   string
		Radius int
	}

	neptune := planet{Name: "Neptune", Radius: 24622000}
	mars := planet{Name: "Mars", Radius: 3389500}
	sameRadiusAsMars := planet{Name: "Same Radius as Mars", Radius: 3389500}

	planets := []planet{mars, neptune, sameRadiusAsMars}
	uniquePlanets := UniqueBy(planets, func(planet planet) int {
		return planet.Radius
	})
	require.Len(t, uniquePlanets, 2)
	assert.Equal(t, mars, uniquePlanets[0])
	assert.Equal(t, neptune, uniquePlanets[1])

	uniquePlanetsRef := UniqueByRef(planets, func(planet *planet) int {
		return planet.Radius
	})
	require.Len(t, uniquePlanetsRef, 2)
	assert.Equal(t, mars, uniquePlanetsRef[0])
	assert.Equal(t, neptune, uniquePlanetsRef[1])

	uniquePlanetsSequence := UniqueBySequence(slices.Values(planets), func(planet planet) int {
		return planet.Radius
	})
	require.Len(t, uniquePlanetsSequence, 2)
	assert.Equal(t, mars, uniquePlanetsSequence[0])
	assert.Equal(t, neptune, uniquePlanetsSequence[1])

	uniquePlanetsRefSequence := UniqueByRefSequence(slices.Values(planets), func(planet *planet) int {
		return planet.Radius
	})
	require.Len(t, uniquePlanetsRefSequence, 2)
	assert.Equal(t, mars, uniquePlanetsRefSequence[0])
	assert.Equal(t, neptune, uniquePlanetsRefSequence[1])
}
