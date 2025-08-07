// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"
)

type DB interface {
	WithTx(ctx context.Context, opts interface{}, fn func(context.Context, interface{}) error) error
}

type TransactionAdapter interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...interface{}) interface{}
}

func DbxGood(ctx context.Context, db DB) error {
	return db.WithTx(ctx, nil, func(ctx context.Context, adapter TransactionAdapter) error {
		// Correct: using adapter parameter
		err := adapter.ExecContext(ctx, "INSERT INTO test VALUES (?)", "value")
		if err != nil {
			return err
		}

		row := adapter.QueryRowContext(ctx, "SELECT id FROM test WHERE value = ?", "value")
		_ = row
		return nil
	})
}
