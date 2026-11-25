package response

type UploadSermonNoteAttachmentFileResponse struct {
	ID             string `json:"id"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	Status         string `json:"status"`
}
