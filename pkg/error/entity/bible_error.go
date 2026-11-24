package entityerror

import (
	apperror "github.com/bryanfks-dev/selah/pkg/error"
)

func ErrBibleGetVersesFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"bible_get_verses_failed",
		"failed to get bible verses",
		cause,
	)
}

func ErrBibleSearchByTextFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"bible_search_by_text_failed",
		"failed to search bible verses by text",
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

func ErrBibleListAvailableTranslationsFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"bible_list_available_translations_failed",
		"failed to list available bible translations",
		cause,
	)
}

func ErrBibleVerseNotFound(ref string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"bible_verse_not_found",
		"bible verse not found: "+ref,
	)
}
