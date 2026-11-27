package uploadsermonnoteattachmentfile

import (
	"context"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bryanfks-dev/logos/pkg/constant"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"go.uber.org/zap"
)

type UploadSermonNoteAttachmentFileInput struct {
	AttachmentID   string
	AttachmentHash string
	File           io.Reader
}

type UploadSermonNoteAttachmentFileResult struct {
	AttachmentID             string
	AttachmentChecksumSHA256 string
	Status                   constant.UploadStatus
}

type writeAndVerifyTempFileOption struct {
	attachment   *entity.SermonNoteAttachment
	expectedHash string
	file         io.Reader
	tmpPath      string
}

type completeUploadOption struct {
	attachment   *entity.SermonNoteAttachment
	computedHash string
	finalPath    string
	writtenBytes int64
}

func (u *UploadSermonNoteAttachmentFileUsecase) Call(
	ctx context.Context,
	input *UploadSermonNoteAttachmentFileInput,
) (*UploadSermonNoteAttachmentFileResult, error) {
	u.log.Info("upload sermon note attachment file started")

	attachment, hash, err := u.validateAndLoadAttachment(ctx, input)
	if err != nil {
		return nil, err
	}

	if attachment.IsUploadedWithValidHash(hash) {
		u.log.Debug(
			"attachment already uploaded with matching hash",
			zap.String("attachment_id", attachment.ID),
		)

		result := &UploadSermonNoteAttachmentFileResult{
			AttachmentID:             attachment.ID,
			AttachmentChecksumSHA256: attachment.ChecksumSHA256,
			Status:                   constant.UploadStatusStored,
		}

		u.log.Info(
			"upload sermon note attachment file completed",
			zap.String("attachment_id", result.AttachmentID),
			zap.String("status", string(result.Status)),
		)

		return result, nil
	}

	if err := u.markProcessing(ctx, attachment); err != nil {
		u.log.Error(
			"failed to mark attachment as processing",
			zap.Error(err),
			zap.String("attachment_id", attachment.ID),
		)

		return nil, err
	}

	tmpPath := filepath.Join(u.appPath, ".tmp", attachment.ID+".part")
	writeOpt := &writeAndVerifyTempFileOption{
		attachment:   attachment,
		expectedHash: hash,
		file:         input.File,
		tmpPath:      tmpPath,
	}

	computedHash, writtenBytes, err := u.writeAndVerifyTempFile(ctx, writeOpt)
	if err != nil {
		return nil, err
	}

	finalFilePath, err := u.moveTempToFinalPath(ctx, attachment, tmpPath)
	if err != nil {
		return nil, err
	}

	completeOpt := &completeUploadOption{
		attachment:   attachment,
		computedHash: computedHash,
		finalPath:    finalFilePath,
		writtenBytes: writtenBytes,
	}

	result, err := u.completeUpload(ctx, completeOpt)
	if err != nil {
		return nil, err
	}

	u.log.Info(
		"upload sermon note attachment file completed",
		zap.String("attachment_id", result.AttachmentID),
		zap.String("status", string(result.Status)),
	)

	return result, nil
}

func (u *UploadSermonNoteAttachmentFileUsecase) validateAndLoadAttachment(
	ctx context.Context,
	input *UploadSermonNoteAttachmentFileInput,
) (*entity.SermonNoteAttachment, string, error) {
	if input == nil || strings.TrimSpace(input.AttachmentID) == "" ||
		input.File == nil {
		u.log.Debug("attachment upload rejected due to invalid file input")

		return nil, "", entityerror.ErrSermonNoteAttachmentInvalidFile()
	}

	attachment, err := u.sermonNoteRepository.FindAttachmentByID(
		ctx,
		input.AttachmentID,
	)
	if err != nil {
		u.log.Error(
			"failed to load sermon note attachment",
			zap.Error(err),
			zap.String("attachment_id", input.AttachmentID),
		)

		return nil, "", err
	}

	hash := strings.ToLower(strings.TrimSpace(input.AttachmentHash))
	if hash == "" {
		u.log.Debug(
			"attachment upload rejected due to empty content hash",
			zap.String("attachment_id", attachment.ID),
		)

		return nil, "", entityerror.ErrSermonNoteAttachmentInvalidHash()
	}

	return attachment, hash, nil
}

func (u *UploadSermonNoteAttachmentFileUsecase) markProcessing(
	ctx context.Context,
	attachment *entity.SermonNoteAttachment,
) error {
	attachment.Status = constant.UploadStatusProcessing
	attachment.UpdatedAt = time.Now().UTC()

	return u.sermonNoteRepository.UpdateAttachment(ctx, attachment)
}

func (u *UploadSermonNoteAttachmentFileUsecase) writeAndVerifyTempFile(
	ctx context.Context,
	opt *writeAndVerifyTempFileOption,
) (string, int64, error) {
	computedHash, writtenBytes, err := u.fsManager.WriteTempAndHashTempFile(
		opt.tmpPath,
		opt.file,
	)
	if err != nil {
		u.resetStatusStored(ctx, opt.attachment)
		u.log.Error(
			"failed to stream attachment to temp file",
			zap.Error(err),
			zap.String("attachment_id", opt.attachment.ID),
		)

		return "", 0, err
	}

	if !u.isHashValid(
		computedHash,
		opt.expectedHash,
		opt.attachment.ChecksumSHA256,
	) {
		_ = u.fsManager.RemoveFile(opt.tmpPath)
		u.resetStatusStored(ctx, opt.attachment)
		u.log.Debug(
			"attachment upload rejected due to hash mismatch",
			zap.String("attachment_id", opt.attachment.ID),
		)

		return "", 0, entityerror.ErrSermonNoteAttachmentInvalidHash()
	}

	return computedHash, writtenBytes, nil
}

func (u *UploadSermonNoteAttachmentFileUsecase) moveTempToFinalPath(
	ctx context.Context,
	attachment *entity.SermonNoteAttachment,
	tmpPath string,
) (string, error) {
	finalFilePath := u.buildFinalFilePath(
		attachment.ID,
		attachment.FileName,
		attachment.CreatedAt,
	)

	if err := u.fsManager.MoveAtomic(tmpPath, finalFilePath); err != nil {
		_ = u.fsManager.RemoveFile(tmpPath)
		u.resetStatusStored(ctx, attachment)
		u.log.Error(
			"failed to move attachment to final path",
			zap.Error(err),
			zap.String("attachment_id", attachment.ID),
		)

		return "", err
	}

	return finalFilePath, nil
}

func (u *UploadSermonNoteAttachmentFileUsecase) completeUpload(
	ctx context.Context,
	opt *completeUploadOption,
) (*UploadSermonNoteAttachmentFileResult, error) {
	opt.attachment.StoragePath = opt.finalPath
	opt.attachment.Status = constant.UploadStatusUploaded
	opt.attachment.UpdatedAt = time.Now().UTC()
	if opt.attachment.FileSize == 0 && opt.writtenBytes > 0 {
		opt.attachment.FileSize = opt.writtenBytes
	}

	if err := u.sermonNoteRepository.UpdateAttachment(
		ctx,
		opt.attachment,
	); err != nil {
		u.log.Error(
			"failed to complete attachment upload",
			zap.Error(err),
			zap.String("attachment_id", opt.attachment.ID),
		)

		return nil, err
	}

	result := &UploadSermonNoteAttachmentFileResult{
		AttachmentID:             opt.attachment.ID,
		AttachmentChecksumSHA256: opt.computedHash,
		Status:                   constant.UploadStatusUploaded,
	}

	return result, nil
}

func (u *UploadSermonNoteAttachmentFileUsecase) isHashValid(
	computedHash string,
	headerHash string,
	storedHash string,
) bool {
	if !strings.EqualFold(computedHash, headerHash) {
		return false
	}

	if storedHash != "" && !strings.EqualFold(computedHash, storedHash) {
		return false
	}

	return true
}

func (u *UploadSermonNoteAttachmentFileUsecase) buildFinalFilePath(
	id string,
	fileName string,
	createdAt time.Time,
) string {
	refTime := createdAt.UTC()
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}

	year := strconv.Itoa(refTime.Year())
	month := strconv.Itoa(int(refTime.Month()))
	ext := filepath.Ext(fileName)

	return filepath.Join(u.appPath, "notes", year, month, id+ext)
}

func (u *UploadSermonNoteAttachmentFileUsecase) resetStatusStored(
	ctx context.Context,
	attachment *entity.SermonNoteAttachment,
) {
	attachment.Status = constant.UploadStatusStored
	attachment.UpdatedAt = time.Now().UTC()
	_ = u.sermonNoteRepository.UpdateAttachment(ctx, attachment)
}
