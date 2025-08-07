// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"

	"storj.io/storj/shared/dbutil/txutil"
	"storj.io/storj/shared/tagsql"
)

func Good(ctx context.Context, db tagsql.DB) error {
	return txutil.WithTx(ctx, db, nil, func(ctx context.Context, tx tagsql.Tx) error {
		// Correct: using tx parameter
		_, err := tx.ExecContext(ctx, "INSERT INTO test VALUES (?)", "value")
		if err != nil {
			return err
		}

		row := tx.QueryRowContext(ctx, "SELECT id FROM test WHERE value = ?", "value")
		var id int
		return row.Scan(&id)
	})
}
