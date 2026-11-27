package util_test

import (
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDatetimeParam(t *testing.T) {
	t.Run("parses RFC3339 datetime", func(t *testing.T) {
		parsed, err := util.ParseDatetimeParam("2026-01-01T12:30:00Z")
		require.NoError(t, err)
		assert.Equal(
			t,
			time.Date(2026, 1, 1, 12, 30, 0, 0, time.UTC),
			parsed.UTC(),
		)
	})

	t.Run("parses RFC3339Nano datetime", func(t *testing.T) {
		parsed, err := util.ParseDatetimeParam("2026-01-01T12:30:00.123456Z")
		require.NoError(t, err)
		assert.False(t, parsed.IsZero())
	})

	t.Run(
		"parses URL-encoded RFC3339 datetime with offset",
		func(t *testing.T) {
			parsed, err := util.ParseDatetimeParam(
				"2026-01-01T00%3A00%3A00%2B07%3A00",
			)
			require.NoError(t, err)
			assert.Equal(
				t,
				time.Date(2025, 12, 31, 17, 0, 0, 0, time.UTC),
				parsed.UTC(),
			)
		},
	)

	t.Run("parses YYYY-MM-DD date", func(t *testing.T) {
		parsed, err := util.ParseDatetimeParam("2026-01-01")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), parsed)
	})

	t.Run("rejects empty or invalid datetime params", func(t *testing.T) {
		cases := []string{
			"",
			"0",
			"not-a-timestamp",
			"%ZZ",
		}

		for _, tc := range cases {
			_, err := util.ParseDatetimeParam(tc)
			require.Error(t, err, "expected error for %q", tc)
		}
	})
}

func TestParseUUIDParam(t *testing.T) {
	t.Run("parses valid UUID", func(t *testing.T) {
		expected := uuid.NewV7()
		parsed, err := util.ParseUUIDParam(expected.String())
		require.NoError(t, err)
		assert.Equal(t, expected, parsed)
	})

	t.Run("rejects empty, nil, or invalid UUID params", func(t *testing.T) {
		cases := []string{
			"",
			"   ",
			"not-a-uuid",
			uuid.Nil().String(),
		}

		for _, tc := range cases {
			parsed, err := util.ParseUUIDParam(tc)
			require.Error(t, err, "expected error for %q", tc)
			assert.Equal(t, uuid.Nil(), parsed)
		}
	})
}
