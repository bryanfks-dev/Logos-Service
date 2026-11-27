package params

import (
	"uuid"

	"github.com/bryanfks-dev/logos/internal/model/param/validator"
	"github.com/bryanfks-dev/logos/pkg/util"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type DownloadSermonNoteAttachmentFileParam struct {
	AttachmentID string `param:"id"`
}

func (p DownloadSermonNoteAttachmentFileParam) ParsedAttachmentID() uuid.UUID {
	parsed, _ := util.ParseUUIDParam(p.AttachmentID)

	return parsed
}

func (p DownloadSermonNoteAttachmentFileParam) Validate() error {
	uuidRule := validation.By(validator.ValidateUUIDParam)
	idField := validation.Field(
		&p.AttachmentID,
		validation.Required,
		uuidRule,
	)

	return validation.ValidateStruct(&p, idField)
}
