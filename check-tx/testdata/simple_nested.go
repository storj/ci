// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"

	"storj.io/storj/shared/dbutil/txutil"
	"storj.io/storj/shared/tagsql"
)

type DbWrapper struct {
	db tagsql.DB
}

func SimpleNested(ctx context.Context, wrapper DbWrapper) error {
	return txutil.WithTx(ctx, wrapper.db, nil, func(ctx context.Context, tx tagsql.Tx) error {
		// BAD: Using wrapper.db instead of tx
		_, err := wrapper.db.ExecContext(ctx, "INSERT INTO test VALUES (?)", "value")
		return err
	})
}
