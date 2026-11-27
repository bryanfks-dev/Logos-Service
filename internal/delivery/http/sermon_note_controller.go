package http

import (
	"mime"
	nethttp "net/http"
	"strings"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	_ "github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/builder"
	"github.com/bryanfks-dev/logos/internal/model/converter"
	params "github.com/bryanfks-dev/logos/internal/model/param"
	"github.com/bryanfks-dev/logos/internal/model/response"
	downloadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/download_sermon_note_attachment_file"
	uploadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/upload_sermon_note_attachment_file"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type SermonNoteController struct {
	uploadAttachmentFileUsecase   *uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecase
	downloadAttachmentFileUsecase *downloadsermonnoteattachmentfile.DownloadSermonNoteAttachmentFileUsecase
	log                           *zap.Logger
}

func NewSermonNoteController(
	uploadAttachmentFileUsecase *uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecase,
	downloadAttachmentFileUsecase *downloadsermonnoteattachmentfile.DownloadSermonNoteAttachmentFileUsecase,
	log *zap.Logger,
) *SermonNoteController {
	return &SermonNoteController{
		uploadAttachmentFileUsecase:   uploadAttachmentFileUsecase,
		downloadAttachmentFileUsecase: downloadAttachmentFileUsecase,
		log:                           log,
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
		AttachmentID:   params.ParsedAttachmentID(),
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

// DownloadAttachmentFile godoc
// @Summary      Download attachment binary file
// @Description  Stream stored binary attachment file with content headers
// @Tags         API / Sermon Note / Attachments
// @Produce      octet-stream
// @Produce      json
// @Param        id path string true "Attachment UUID"
// @Success      200 {file} binary "Raw binary attachment file stream"
// @Failure      400 {object} model.APIResponse[any]
// @Failure      404 {object} model.APIResponse[any]
// @Failure      422 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/notes/{id}/file [get]
func (c *SermonNoteController) DownloadAttachmentFile(
	ctx *echo.Context,
) error {
	reqID := correlation.GetRequestID(ctx)

	params := new(params.DownloadSermonNoteAttachmentFileParam)
	if err := echo.BindPathValues(ctx, params); err != nil {
		c.log.Warn(
			"failed to bind attachment download params",
			zap.Error(err),
			zap.String("request_id", reqID),
		)

		return apperror.NewErrRequestParamsMalformed()
	}

	if err := params.Validate(); err != nil {
		c.log.Warn(
			"failed to validate attachment download params",
			zap.Error(err),
			zap.String("attachment_id", params.AttachmentID),
			zap.String("request_id", reqID),
		)

		return err
	}

	attachmentID := params.ParsedAttachmentID()

	result, err := c.downloadAttachmentFileUsecase.Call(
		ctx.Request().Context(),
		attachmentID,
	)
	if err != nil {
		return err
	}
	defer result.File.Close()

	contentType := strings.TrimSpace(result.MimeType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName := strings.TrimSpace(result.FileName)
	if fileName == "" {
		fileName = attachmentID.String()
	}

	ctx.Response().Header().Set(echo.HeaderContentType, contentType)

	dispParams := map[string]string{"filename": fileName}
	disposition := mime.FormatMediaType("attachment", dispParams)
	if disposition != "" {
		ctx.Response().Header().Set(echo.HeaderContentDisposition, disposition)
	}

	if result.ChecksumSHA256 != "" {
		ctx.Response().Header().Set("X-Content-Hash", result.ChecksumSHA256)
	}

	nethttp.ServeContent(
		ctx.Response(),
		ctx.Request(),
		fileName,
		result.UpdatedAt,
		result.File,
	)

	return nil
}
