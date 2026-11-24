package dbcorrelation

import (
	"context"

	"gorm.io/gorm"
)

const transactionKey = "tx"

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, transactionKey, tx)
}

func GetTx(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return nil
	}
	if tx, ok := ctx.Value(transactionKey).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return nil
}
