package converter

import (
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/response"
)

func SermonNoteAttachmentUploadResultToResponse(
	result *model.UploadSermonNoteAttachmentFileResult,
) *response.UploadSermonNoteAttachmentFileResponse {
	if result == nil {
		return nil
	}

	return &response.UploadSermonNoteAttachmentFileResponse{
		ID:             result.ID,
		ChecksumSHA256: result.ChecksumSHA256,
		Status:         result.Status,
	}
}
