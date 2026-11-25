package correlation

import (
	"context"

	"gorm.io/gorm"
)

const transactionKey string = "tx"

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionKey, tx)
}

func GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transactionKey).(*gorm.DB); ok && tx != nil {
		return tx
	}

	return nil
}
