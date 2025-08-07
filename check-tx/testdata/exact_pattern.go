// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.
package testdata

import (
	"context"

	"storj.io/storj/satellite/satellitedb/dbx"
	"storj.io/storj/shared/tagsql"
)

type rollupToDelete struct {
	ProjectID     []byte
	BucketName    []byte
	IntervalStart interface{}
	Action        int64
}

func withRows(rows tagsql.Rows, err error) func(func(tagsql.Rows) error) error {
	return func(fn func(tagsql.Rows) error) error {
		if err != nil {
			return err
		}
		defer rows.Close()
		return fn(rows)
	}
}

type DatabaseWithNested struct {
	db interface {
		WithTx(ctx context.Context, fn func(context.Context, *dbx.Tx) error) error
		QueryContext(ctx context.Context, query string, args ...interface{}) (tagsql.Rows, error)
		ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error)
	}
}

// This matches the exact pattern from the user's example
func ExactPattern(ctx context.Context, db DatabaseWithNested, before interface{}, batchSize int) error {
	query := "SELECT project_id, bucket_name, interval_start, action FROM bucket_bandwidth_rollups ORDER BY interval_start LIMIT ?"
	var rowCount int64
	var archivedCount int

	err := db.db.WithTx(ctx, func(ctx context.Context, tx *dbx.Tx) error {
		return withRows(db.db.QueryContext(ctx, query, before, batchSize))(func(rows tagsql.Rows) error {
			var toDelete []rollupToDelete
			for rows.Next() {
				var rollup rollupToDelete
				if err := rows.Scan(&rollup.ProjectID, &rollup.BucketName, &rollup.IntervalStart, &rollup.Action); err != nil {
					return err
				}
				toDelete = append(toDelete, rollup)
			}

			res, err := db.db.ExecContext(ctx, `
				DELETE FROM bucket_bandwidth_rollups
					WHERE STRUCT<ProjectID BYTES, BucketName BYTES, IntervalStart TIMESTAMP, Action INT64>(project_id, bucket_name, interval_start, action) IN UNNEST(?)`,
				toDelete)
			if err != nil {
				return err
			}

			_ = res
			_ = rowCount
			_ = archivedCount

			return nil
		})
	})
	return err
}
