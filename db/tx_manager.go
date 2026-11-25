package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bryanfks-dev/selah/db/correlation"
	"gorm.io/gorm"
)

type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) Do(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	return m.DoWithOptions(ctx, nil, fn)
}

func (m *TxManager) DoWithOptions(
	ctx context.Context,
	opts *sql.TxOptions,
	fn func(ctx context.Context) error,
) error {
	if m == nil || m.db == nil {
		return errors.New("db is nil")
	}
	if fn == nil {
		return nil
	}

	if tx := correlation.Tx(ctx); tx != nil {
		return fn(ctx)
	}

	var txOpts []*sql.TxOptions
	if opts != nil {
		txOpts = append(txOpts, opts)
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		extendedCtx := correlation.WithTx(ctx, tx)
		return fn(extendedCtx)
	}, txOpts...)
}
