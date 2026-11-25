package params_test

import (
	"testing"

	"github.com/bryanfks-dev/logos/internal/model/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncPullChangesParams_Validate(t *testing.T) {
	t.Run("requires since", func(t *testing.T) {
		p := params.SyncPullChangesParam{}
		err := p.Validate()
		require.Error(t, err)
	})

	t.Run("accepts valid datetime formats", func(t *testing.T) {
		cases := []string{
			"2026-01-01",
			"2026-01-01T00:00:00Z",
			"2026-01-01T00:00:00.123456Z",
			"2026-01-01T00%3A00%3A00%2B07%3A00",
		}

		for _, tc := range cases {
			p := params.SyncPullChangesParam{Since: tc}
			assert.NoError(t, p.Validate(), "expected valid for %s", tc)
		}
	})

	t.Run("rejects invalid datetime formats", func(t *testing.T) {
		cases := []string{
			"0",
			"not-a-timestamp",
			"%ZZ",
		}

		for _, tc := range cases {
			p := params.SyncPullChangesParam{Since: tc}
			require.Error(t, p.Validate(), "expected error for %s", tc)
		}
	})
}
