package constant

type UploadStatus string

const (
	UploadStatusStored     UploadStatus = "stored"
	UploadStatusProcessing UploadStatus = "processing"
	UploadStatusUploaded   UploadStatus = "uploaded"
)
