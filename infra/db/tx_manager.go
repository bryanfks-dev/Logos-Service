package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"gorm.io/gorm"
)

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
	DoWithOptions(
		ctx context.Context,
		opts *sql.TxOptions,
		fn func(ctx context.Context) error,
	) error
}

type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) TxManager {
	return &txManager{db: db}
}

func (m *txManager) Do(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	return m.DoWithOptions(ctx, nil, fn)
}

func (m *txManager) DoWithOptions(
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

	if tx := correlation.GetTx(ctx); tx != nil {
		return fn(ctx)
	}

	var txOpts []*sql.TxOptions
	if opts != nil {
		txOpts = append(txOpts, opts)
	}

	hookCtx, hookFn := WithAfterCommit(ctx)

	txFn := func(tx *gorm.DB) error {
		extendedCtx := correlation.WithTx(hookCtx, tx)
		return fn(extendedCtx)
	}

	err := m.db.WithContext(ctx).Transaction(txFn, txOpts...)
	if err != nil {
		return err
	}

	return hookFn()
}
