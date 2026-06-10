// Package database provides shared database utilities including transaction management.
package database

import (
	"context"

	"gorm.io/gorm"
)

// TXKey is the context key used to store the current transaction.
type TXKey struct{}

// TransactionManager runs functions inside database transactions.
type TransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionManager struct {
	db *gorm.DB
}

// NewTransactionManager returns a TransactionManager backed by the given *gorm.DB.
func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManager{db: db}
}

// Run executes fn within a database transaction. If fn returns an error,
// the transaction is rolled back. Otherwise it is committed.
// The transaction *gorm.DB is stored in the context and can be retrieved
// via TXFromContext.
func (tm *transactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, TXKey{}, tx)
		return fn(ctx)
	})
}

// TXFromContext returns the transaction stored in ctx, or falls back to db.
// Repositories use this to automatically participate in an active transaction.
func TXFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(TXKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}
