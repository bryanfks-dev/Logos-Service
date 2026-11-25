package constant

type UploadStatus string

const (
	UploadStatusNotUploaded UploadStatus = "not_uploaded"
	UploadStatusProcessing  UploadStatus = "processing"
	UploadStatusUploaded    UploadStatus = "uploaded"
)
