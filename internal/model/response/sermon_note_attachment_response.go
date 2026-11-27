package response

import "github.com/bryanfks-dev/logos/pkg/constant"

type UploadSermonNoteAttachmentFileResponse struct {
	ID             string                `json:"id"`
	ChecksumSHA256 string                `json:"checksum_sha256"`
	Status         constant.UploadStatus `json:"status"`
}
