package correlation

import (
	"context"

	"gorm.io/gorm"
)

const transactionKey = "tx"

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionKey, tx)
}

func Tx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transactionKey).(*gorm.DB); ok && tx != nil {
		return tx
	}

	return nil
}

func GetTx(ctx context.Context) *gorm.DB {
	return Tx(ctx)
}
