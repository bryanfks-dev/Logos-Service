package params_test

import (
	"testing"
	"uuid"

	params "github.com/bryanfks-dev/logos/internal/model/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadSermonNoteAttachmentFileParam_Validate(t *testing.T) {
	t.Run("rejects empty, nil, or invalid attachment id", func(t *testing.T) {
		cases := []string{
			"",
			"   ",
			"not-a-uuid",
			uuid.Nil().String(),
		}

		for _, tc := range cases {
			p := params.UploadSermonNoteAttachmentFileParam{
				AttachmentID: tc,
			}
			require.Error(t, p.Validate(), "expected error for %q", tc)
			assert.Equal(t, uuid.Nil(), p.ParsedAttachmentID())
		}
	})

	t.Run("accepts and parses valid UUID attachment id", func(t *testing.T) {
		expected := uuid.NewV7()
		p := params.UploadSermonNoteAttachmentFileParam{
			AttachmentID: expected.String(),
		}
		assert.NoError(t, p.Validate())
		assert.Equal(t, expected, p.ParsedAttachmentID())
	})
}
