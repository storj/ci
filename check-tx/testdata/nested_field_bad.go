// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"
)

type NestedDB struct {
	db interface {
		WithTx(ctx context.Context, fn func(context.Context, *Tx) error) error
		ExecContext(ctx context.Context, query string, args ...interface{}) error
		QueryContext(ctx context.Context, query string, args ...interface{}) error
	}
}

type Tx struct{}

func NestedFieldBad(ctx context.Context, db NestedDB) error {
	return db.db.WithTx(ctx, func(ctx context.Context, tx *Tx) error {
		// BAD: Using db.db.ExecContext instead of tx
		_, err := db.db.ExecContext(ctx, "DELETE FROM table WHERE id = ?", 1)
		if err != nil {
			return err
		}

		// BAD: Using db.db.QueryContext instead of tx
		_, err = db.db.QueryContext(ctx, "SELECT * FROM table")
		return err
	})
}
