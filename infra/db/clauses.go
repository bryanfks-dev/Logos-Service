package db

import (
	"context"

	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func WithRowLock(ctx context.Context, query *gorm.DB) *gorm.DB {
	if query == nil {
		return nil
	}

	if correlation.GetTx(ctx) != nil {
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	return query
}
