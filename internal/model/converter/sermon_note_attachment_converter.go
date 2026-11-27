package converter

import (
	"github.com/bryanfks-dev/logos/internal/model/response"
	uploadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/upload_sermon_note_attachment_file"
)

func SermonNoteAttachmentUploadResultToResponse(
	result *uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileResult,
) *response.UploadSermonNoteAttachmentFileResponse {
	if result == nil {
		return nil
	}

	return &response.UploadSermonNoteAttachmentFileResponse{
		ID:             result.AttachmentID,
		ChecksumSHA256: result.AttachmentChecksumSHA256,
		Status:         result.Status,
	}
}
