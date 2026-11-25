package util_test

import (
	"testing"

	"github.com/bryanfks-dev/logos/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestCoalesceSlice(t *testing.T) {
	t.Run("returns primary when non-empty", func(t *testing.T) {
		primary := []string{"a", "b"}
		fallback := []string{"c"}

		assert.Equal(t, primary, util.CoalesceSlice(primary, fallback))
	})

	t.Run("returns fallback when primary is nil or empty", func(t *testing.T) {
		fallback := []int{1, 2}
		emptyInt := []int{}

		assert.Equal(t, fallback, util.CoalesceSlice(nil, fallback))
		assert.Equal(t, fallback, util.CoalesceSlice(emptyInt, fallback))
	})

	t.Run(
		"returns nil or empty fallback when both are empty",
		func(t *testing.T) {
			emptyPrimary := []string{}
			emptyFallback := []string{}

			assert.Nil(t, util.CoalesceSlice[string](nil, nil))
			assert.Empty(t, util.CoalesceSlice(emptyPrimary, emptyFallback))
		},
	)
}

func TestNonNilSlice(t *testing.T) {
	t.Run(
		"returns empty non-nil slice when input is nil or empty",
		func(t *testing.T) {
			nilResult := util.NonNilSlice[string](nil)
			assert.NotNil(t, nilResult)
			assert.Empty(t, nilResult)

			emptyInput := []string{}
			emptyResult := util.NonNilSlice(emptyInput)
			assert.NotNil(t, emptyResult)
			assert.Empty(t, emptyResult)
		},
	)

	t.Run("returns original slice when non-empty", func(t *testing.T) {
		items := []int{10, 20}

		assert.Equal(t, items, util.NonNilSlice(items))
	})
}
