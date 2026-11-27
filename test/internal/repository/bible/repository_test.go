package bible_test

import (
	"context"
	"testing"

	"github.com/bryanfks-dev/logos/internal/repository/bible"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBibleTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{},
	)
	require.NoError(t, err)

	err = testDB.AutoMigrate(&entity.BibleVerse{}, &entity.BibleFTS{})
	require.NoError(t, err)

	return testDB
}

func TestBibleRepository(t *testing.T) {
	ctx := context.Background()
	testDB := setupBibleTestDB(t)
	testLog := zap.NewNop()

	repo := bible.NewBibleRepository(testDB, testLog)

	verses := []entity.BibleVerse{
		{
			Translation: "ESV",
			BookID:      1,
			BookName:    "Genesis",
			Chapter:     1,
			Verse:       1,
			Text:        "In the beginning, God created the heavens and the earth.",
		},
		{
			Translation: "ESV",
			BookID:      1,
			BookName:    "Genesis",
			Chapter:     1,
			Verse:       2,
			Text:        "The earth was without form and void.",
		},
		{
			Translation: "NIV",
			BookID:      1,
			BookName:    "Genesis",
			Chapter:     1,
			Verse:       1,
			Text:        "In the beginning God created the heavens and the earth.",
		},
	}

	assert.NoError(t, repo.ImportTranslationPack(ctx, verses))
	assert.NoError(t, repo.ImportTranslationPack(ctx, nil))

	translations, err := repo.ListAvailableTranslations(ctx)
	assert.NoError(t, err)
	assert.Len(t, translations, 2)

	end := 2

	verseQuery := entity.VerseQuery{
		Translation: "ESV",
		BookID:      1,
		Chapter:     1,
		VerseStart:  1,
		VerseEnd:    &end,
	}
	fetched, err := repo.GetVerses(ctx, verseQuery)
	assert.NoError(t, err)
	assert.Len(t, fetched, 2)

	searched, err := repo.SearchByText(ctx, "ESV", "beginning")
	assert.NoError(t, err)
	assert.Len(t, searched, 1)
}
