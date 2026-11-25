package entityerror

import (
	apperror "github.com/bryanfks-dev/logos/pkg/error"
)

func ErrBibleRetrievalFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"bible_retrieval_failed",
		"failed to retrieve bible data",
		cause,
	)
}

func ErrBibleImportTranslationPackFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"bible_import_translation_pack_failed",
		"failed to import bible translation pack",
		cause,
	)
}

func ErrBibleVerseNotFound(ref string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"bible_verse_not_found",
		"bible verse not found: "+ref,
	)
}
