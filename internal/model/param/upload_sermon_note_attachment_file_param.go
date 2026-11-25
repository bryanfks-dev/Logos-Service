package params

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type UploadSermonNoteAttachmentFileParam struct {
	AttachmentID string `param:"id"`
}

func (p UploadSermonNoteAttachmentFileParam) Validate() error {
	idField := validation.Field(&p.AttachmentID, validation.Required)

	return validation.ValidateStruct(&p, idField)
}
