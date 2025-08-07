# Transaction Linter

A static analysis tool that checks that transaction callback functions use the transaction parameter (`tx`) instead of accessing the database directly (`db`).

## Problem

When using transaction wrappers like `txutil.WithTx`, `sqliteutil.WithTx`, or `db.WithTx`, the callback function receives a transaction parameter that should be used for all database operations. Using the original database parameter bypasses the transaction and can lead to data consistency issues.

## Bad Examples

### txutil.WithTx / sqliteutil.WithTx
```go
return txutil.WithTx(ctx, db, nil, func(ctx context.Context, tx tagsql.Tx) error {
    // BAD: Using db instead of tx
    _, err := db.ExecContext(ctx, "INSERT INTO users VALUES (?)", "john")
    return err
})
```

### db.WithTx (Metabase pattern)
```go
return db.ChooseAdapter().WithTx(ctx, opts, func(ctx context.Context, adapter TransactionAdapter) error {
    // BAD: Using db instead of adapter
    err := db.ExecContext(ctx, "INSERT INTO users VALUES (?)", "john")
    return err
})
```

### Nested field access
```go
return db.db.WithTx(ctx, func(ctx context.Context, tx *dbx.Tx) error {
    // BAD: Using db.db instead of tx
    _, err := db.db.ExecContext(ctx, "DELETE FROM table WHERE id = ?", 1)
    return err
})
```

## Good Examples

### txutil.WithTx / sqliteutil.WithTx
```go
return txutil.WithTx(ctx, db, nil, func(ctx context.Context, tx tagsql.Tx) error {
    // GOOD: Using tx parameter
    _, err := tx.ExecContext(ctx, "INSERT INTO users VALUES (?)", "john")
    return err
})
```

### db.WithTx (Metabase pattern)
```go
return db.ChooseAdapter().WithTx(ctx, opts, func(ctx context.Context, adapter TransactionAdapter) error {
    // GOOD: Using adapter parameter
    err := adapter.ExecContext(ctx, "INSERT INTO users VALUES (?)", "john")
    return err
})
```

### Nested field access
```go
return db.db.WithTx(ctx, func(ctx context.Context, tx *dbx.Tx) error {
    // GOOD: Using tx parameter
    _, err := tx.ExecContext(ctx, "DELETE FROM table WHERE id = ?", 1)
    return err
})
```

## Usage

Check a single file:
```bash
./check-tx -path path/to/file.go
```

Check a directory:
```bash
./check-tx -path path/to/directory
```

Verbose output:
```bash
./check-tx -path . -v
```

## Detected Methods

The linter checks for incorrect usage of these database methods:
- `Exec`, `ExecContext`
- `Query`, `QueryContext`  
- `QueryRow`, `QueryRowContext`
- `Prepare`, `PrepareContext`

## Supported Transaction Patterns

The linter recognizes these WithTx patterns:

1. **txutil.WithTx**: `txutil.WithTx(ctx, db, opts, func(ctx, tx) error)`
2. **sqliteutil.WithTx**: `sqliteutil.WithTx(ctx, db, func(ctx, tx) error)`  
3. **db.WithTx**: `db.WithTx(ctx, opts, func(ctx, adapter) error)` (Metabase/DBX pattern)
4. **Method chains**: `db.ChooseAdapter().WithTx(ctx, opts, func(ctx, adapter) error)`

## Integration

The tool returns exit code 1 if issues are found, making it suitable for CI/CD pipelines.

Add to your linting pipeline:
```bash
make llint && ./check-tx -path .
```