package usecase

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bryanfks-dev/logos/infra/fs"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/repository"
	"github.com/bryanfks-dev/logos/pkg/constant"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type UploadSermonNoteAttachmentFileUsecase struct {
	sermonNoteRepository repository.SermonNoteRepository
	fsManager            fs.Manager
	appPath              string
	log                  *zap.Logger
}

type UploadSermonNoteAttachmentFileUsecaseOption struct {
	SermonNoteRepository repository.SermonNoteRepository
	FsManager            fs.Manager
	Config               *viper.Viper
	Log                  *zap.Logger
}

func NewUploadSermonNoteAttachmentFileUsecase(
	opt *UploadSermonNoteAttachmentFileUsecaseOption,
) *UploadSermonNoteAttachmentFileUsecase {
	appPath := ""
	if opt.Config != nil {
		appPath = opt.Config.GetString("app.path")
	}

	return &UploadSermonNoteAttachmentFileUsecase{
		sermonNoteRepository: opt.SermonNoteRepository,
		fsManager:            opt.FsManager,
		appPath:              appPath,
		log:                  opt.Log,
	}
}

func (u *UploadSermonNoteAttachmentFileUsecase) Call(
	ctx context.Context,
	input *model.UploadSermonNoteAttachmentFileInput,
) (*model.UploadSermonNoteAttachmentFileResult, error) {
	if input == nil || strings.TrimSpace(input.ID) == "" || input.File == nil {
		return nil, entityerror.ErrSermonNoteAttachmentInvalidFile()
	}

	attachment, err := u.sermonNoteRepository.FindAttachmentByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	contentHash := strings.ToLower(strings.TrimSpace(input.ContentHash))
	if contentHash == "" {
		return nil, entityerror.ErrSermonNoteAttachmentInvalidHash()
	}

	if attachment.IsUploadedWithHash(contentHash) {
		result := &model.UploadSermonNoteAttachmentFileResult{
			ID:             attachment.ID,
			ChecksumSHA256: attachment.ChecksumSHA256,
			Status:         "stored",
		}

		return result, nil
	}

	attachment.Status = constant.UploadStatusProcessing
	attachment.UpdatedAt = time.Now().UTC()
	if err := u.sermonNoteRepository.UpdateAttachment(ctx, attachment); err != nil {
		return nil, err
	}

	tmpPath := filepath.Join(u.appPath, ".tmp", attachment.ID+".part")

	computedHash, writtenBytes, err := u.fsManager.WriteTempAndHash(
		tmpPath,
		input.File,
	)
	if err != nil {
		u.resetStatusNotUploaded(ctx, attachment)
		u.log.Error(
			"failed to stream attachment to temp file",
			zap.Error(err),
			zap.String("attachment_id", attachment.ID),
		)

		return nil, err
	}

	if !u.isHashValid(computedHash, contentHash, attachment.ChecksumSHA256) {
		_ = u.fsManager.RemoveFile(tmpPath)
		u.resetStatusNotUploaded(ctx, attachment)

		return nil, entityerror.ErrSermonNoteAttachmentInvalidHash()
	}

	finalPath := u.buildFinalPath(
		attachment.ID,
		attachment.FileName,
		attachment.CreatedAt,
	)

	if err := u.fsManager.MoveAtomic(tmpPath, finalPath); err != nil {
		_ = u.fsManager.RemoveFile(tmpPath)
		u.resetStatusNotUploaded(ctx, attachment)
		u.log.Error(
			"failed to move attachment to final path",
			zap.Error(err),
			zap.String("attachment_id", attachment.ID),
		)

		return nil, err
	}

	attachment.StoragePath = finalPath
	attachment.Status = constant.UploadStatusUploaded
	attachment.UpdatedAt = time.Now().UTC()
	if attachment.FileSize == 0 && writtenBytes > 0 {
		attachment.FileSize = writtenBytes
	}

	if err := u.sermonNoteRepository.UpdateAttachment(ctx, attachment); err != nil {
		return nil, err
	}

	result := &model.UploadSermonNoteAttachmentFileResult{
		ID:             attachment.ID,
		ChecksumSHA256: computedHash,
		Status:         "stored",
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

func (u *UploadSermonNoteAttachmentFileUsecase) buildFinalPath(
	id string,
	fileName string,
	createdAt time.Time,
) string {
	refTime := createdAt.UTC()
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}

	year := strconv.Itoa(refTime.Year())
	ext := filepath.Ext(fileName)

	return filepath.Join(u.appPath, year, id+ext)
}

func (u *UploadSermonNoteAttachmentFileUsecase) resetStatusNotUploaded(
	ctx context.Context,
	attachment *entity.SermonNoteAttachment,
) {
	attachment.Status = constant.UploadStatusNotUploaded
	attachment.UpdatedAt = time.Now().UTC()
	_ = u.sermonNoteRepository.UpdateAttachment(ctx, attachment)
}
