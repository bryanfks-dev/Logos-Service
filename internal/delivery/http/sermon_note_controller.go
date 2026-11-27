package http

import (
	nethttp "net/http"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	_ "github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/builder"
	"github.com/bryanfks-dev/logos/internal/model/converter"
	params "github.com/bryanfks-dev/logos/internal/model/param"
	"github.com/bryanfks-dev/logos/internal/model/response"
	uploadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/upload_sermon_note_attachment_file"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type SermonNoteController struct {
	uploadAttachmentFileUsecase *uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecase
	log                         *zap.Logger
}

func NewSermonNoteController(
	uploadAttachmentFileUsecase *uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecase,
	log *zap.Logger,
) *SermonNoteController {
	return &SermonNoteController{
		uploadAttachmentFileUsecase: uploadAttachmentFileUsecase,
		log:                         log,
	}
}

// UploadAttachmentFile godoc
// @Summary      Upload attachment binary file
// @Description  Stream raw binary attachment file, verify SHA-256 hash, store atomically, and mark uploaded
// @Tags         API / Sermon Note / Attachments
// @Accept       octet-stream
// @Produce      json
// @Param        id path string true "Attachment UUID"
// @Param        X-Content-Hash header string true "Expected SHA-256 hex digest"
// @Param        body body string true "Raw binary attachment file stream" format(binary)
// @Success      200 {object} model.APIResponse[response.UploadSermonNoteAttachmentFileResponse]
// @Failure      400 {object} model.APIResponse[any]
// @Failure      404 {object} model.APIResponse[any]
// @Failure      422 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/attachments/{id}/file [put]
func (c *SermonNoteController) UploadAttachmentFile(ctx *echo.Context) error {
	reqID := correlation.GetRequestID(ctx)

	params := new(params.UploadSermonNoteAttachmentFileParam)
	if err := echo.BindPathValues(ctx, params); err != nil {
		c.log.Warn(
			"failed to bind attachment upload params",
			zap.Error(err),
			zap.String("request_id", reqID),
		)

		return apperror.NewErrRequestParamsMalformed()
	}

	if err := params.Validate(); err != nil {
		c.log.Warn(
			"failed to validate attachment upload params",
			zap.Error(err),
			zap.String("attachment_id", params.AttachmentID),
			zap.String("request_id", reqID),
		)

		return err
	}

	input := &uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
		AttachmentID:   params.AttachmentID,
		AttachmentHash: ctx.Request().Header.Get("X-Content-Hash"),
		File:           ctx.Request().Body,
	}

	result, err := c.uploadAttachmentFileUsecase.Call(
		ctx.Request().Context(),
		input,
	)
	if err != nil {
		return err
	}

	resData := converter.SermonNoteAttachmentUploadResultToResponse(result)
	res := builder.NewAPIResponseBuilder[response.UploadSermonNoteAttachmentFileResponse]().
		SetRequestID(reqID).
		SetData(resData).
		Build()

	return ctx.JSON(nethttp.StatusOK, res)
}
