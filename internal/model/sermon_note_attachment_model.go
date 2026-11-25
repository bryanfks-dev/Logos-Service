package model

import "io"

type UploadSermonNoteAttachmentFileInput struct {
	ID          string
	ContentHash string
	File        io.Reader
}

type UploadSermonNoteAttachmentFileResult struct {
	ID             string
	ChecksumSHA256 string
	Status         string
}
