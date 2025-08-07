// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"
)

type DB2 interface {
	WithTx(ctx context.Context, opts interface{}, fn func(context.Context, interface{}) error) error
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...interface{}) interface{}
}

type TransactionAdapter2 interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...interface{}) interface{}
}

func DbxBad(ctx context.Context, db DB2) error {
	return db.WithTx(ctx, nil, func(ctx context.Context, adapter TransactionAdapter2) error {
		// Incorrect: using db instead of adapter
		err := db.ExecContext(ctx, "INSERT INTO test VALUES (?)", "value")
		if err != nil {
			return err
		}

		// Also incorrect: using db instead of adapter
		row := db.QueryRowContext(ctx, "SELECT id FROM test WHERE value = ?", "value")
		_ = row
		return nil
	})
}
