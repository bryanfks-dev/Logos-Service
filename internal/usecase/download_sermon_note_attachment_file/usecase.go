package downloadsermonnoteattachmentfile

import (
	"context"
	"errors"
	"io"
	"os"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/pkg/entity"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"go.uber.org/zap"
)

type DownloadSermonNoteAttachmentFileResult struct {
	File           io.ReadSeekCloser
	FileName       string
	MimeType       string
	FileSize       int64
	ChecksumSHA256 string
	UpdatedAt      time.Time
}

func (u *DownloadSermonNoteAttachmentFileUsecase) Call(
	ctx context.Context,
	attachmentID uuid.UUID,
) (*DownloadSermonNoteAttachmentFileResult, error) {
	u.log.Info(
		"download sermon note attachment file started",
		zap.String("attachment_id", attachmentID.String()),
	)

	attachment, err := u.loadUploadedAttachment(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	file, info, err := u.openAttachmentFile(attachment)
	if err != nil {
		return nil, err
	}

	fileSize := attachment.FileSize
	if fileSize == 0 {
		fileSize = info.Size()
	}

	updatedAt := attachment.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = info.ModTime().UTC()
	}

	result := &DownloadSermonNoteAttachmentFileResult{
		File:           file,
		FileName:       attachment.FileName,
		MimeType:       attachment.MimeType,
		FileSize:       fileSize,
		ChecksumSHA256: attachment.ChecksumSHA256,
		UpdatedAt:      updatedAt,
	}

	u.log.Info(
		"download sermon note attachment file completed",
		zap.String("attachment_id", attachment.ID),
		zap.Int64("file_size", result.FileSize),
	)

	return result, nil
}

func (u *DownloadSermonNoteAttachmentFileUsecase) loadUploadedAttachment(
	ctx context.Context,
	attachmentID uuid.UUID,
) (*entity.SermonNoteAttachment, error) {
	id := attachmentID.String()
	if attachmentID == uuid.Nil() {
		u.log.Debug("attachment download rejected due to nil attachment id")

		return nil, entityerror.ErrSermonNoteAttachmentNotFound(id)
	}

	attachment, err := u.sermonNoteRepository.FindAttachmentByID(ctx, id)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.IsDomain() {
			return nil, err
		}

		u.log.Error(
			"failed to load sermon note attachment",
			zap.Error(err),
			zap.String("attachment_id", id),
		)

		return nil, err
	}

	if !attachment.IsFileUploaded() {
		u.log.Debug(
			"attachment file is not uploaded yet",
			zap.String("attachment_id", attachment.ID),
			zap.String("status", string(attachment.Status)),
		)

		return nil, entityerror.ErrSermonNoteAttachmentNotFound(attachment.ID)
	}

	return attachment, nil
}

func (u *DownloadSermonNoteAttachmentFileUsecase) openAttachmentFile(
	attachment *entity.SermonNoteAttachment,
) (io.ReadSeekCloser, os.FileInfo, error) {
	file, info, err := u.fsManager.OpenFile(attachment.StoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			u.log.Debug(
				"attachment file not found on storage",
				zap.String("attachment_id", attachment.ID),
				zap.String("storage_path", attachment.StoragePath),
			)

			return nil, nil, entityerror.ErrSermonNoteAttachmentNotFound(
				attachment.ID,
			)
		}

		u.log.Error(
			"failed to open sermon note attachment file",
			zap.Error(err),
			zap.String("attachment_id", attachment.ID),
			zap.String("storage_path", attachment.StoragePath),
		)

		return nil, nil, entityerror.ErrSermonNoteAttachmentRetrievalFailed(err)
	}

	return file, info, nil
}
