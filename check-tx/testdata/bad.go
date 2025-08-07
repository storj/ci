// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"

	"storj.io/storj/shared/dbutil/txutil"
	"storj.io/storj/shared/tagsql"
)

func Bad(ctx context.Context, db tagsql.DB) error {
	return txutil.WithTx(ctx, db, nil, func(ctx context.Context, tx tagsql.Tx) error {
		// Incorrect: using db instead of tx
		_, err := db.ExecContext(ctx, "INSERT INTO test VALUES (?)", "value")
		if err != nil {
			return err
		}

		// Also incorrect: using db instead of tx
		row := db.QueryRowContext(ctx, "SELECT id FROM test WHERE value = ?", "value")
		var id int
		return row.Scan(&id)
	})
}
