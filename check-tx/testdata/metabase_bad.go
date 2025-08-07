// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"
)

type MetabaseDB interface {
	ChooseAdapter() MetabaseAdapter
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	PrecommitConstraint(ctx context.Context, constraint interface{}, adapter interface{}) error
}

type MetabaseAdapter interface {
	WithTx(ctx context.Context, opts interface{}, fn func(context.Context, TransactionAdapter3) error) error
}

type TransactionAdapter3 interface {
	fetchSegmentsForCommit(ctx context.Context, streamID interface{}) error
}

func MetabaseBad(ctx context.Context, db MetabaseDB) error {
	return db.ChooseAdapter().WithTx(ctx, nil, func(ctx context.Context, adapter TransactionAdapter3) error {
		// Incorrect: using db instead of adapter
		err := db.ExecContext(ctx, "INSERT INTO test VALUES (?)", "value")
		if err != nil {
			return err
		}

		// This should be fine since it's using adapter
		return adapter.fetchSegmentsForCommit(ctx, "streamID")
	})
}
