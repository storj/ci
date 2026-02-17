// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

package a

// WithRetry calls fn, potentially multiple times on failure.
func WithRetry(fn func()) { fn() }

// Retrier has methods that retry.
type Retrier struct{}

// ReadWriteTransaction retries the callback on transient failures.
func (r *Retrier) ReadWriteTransaction(fn func(tx int) error) error { return fn(0) }

// ReadTransaction retries the callback on transient failures.
func (r *Retrier) ReadTransaction(fn func(tx int) error) error { return fn(0) }

// WithTx runs fn inside a transaction, retrying on transient failures.
func (r *Retrier) WithTx(fn func(tx int) error) error { return fn(0) }

// Options for transactions.
type Options struct{}

// ReadWriteTransactionWithOptions retries the callback on transient failures.
func (r *Retrier) ReadWriteTransactionWithOptions(fn func(tx int) error, opts Options) (int, error) {
	return 0, fn(0)
}

// CtxReadWriteTransactionWithOptions is like ReadWriteTransactionWithOptions but takes ctx.
func (r *Retrier) CtxReadWriteTransactionWithOptions(ctx interface{}, fn func(ctx interface{}, tx int) error, opts Options) (int, error) {
	return 0, fn(ctx, 0)
}

// QueryResult simulates a query returning multiple values.
func QueryResult(ctx interface{}, tx int, sql string) (int, error) { return tx, nil }

// CollectRow simulates spannerutil.CollectRow.
func CollectRow(v int, err error) (int, error) { return v, err }
