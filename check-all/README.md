# check-all

`check-all` is a multichecker tool that runs multiple static analysis checks in a single pass.

## Usage

```bash
go build -o check-all ./check-all
./check-all [flags] [packages]
```

## Included Analyzers

This tool includes the following analyzers:

- **callsize**: Check method/function calls where large number of bytes are passed
- **deferloop**: Check for defers inside a loop
- **errs**: Check for proper usage of errs package
- **monkitunused**: Check for unfinished calls to mon.Task()(ctx)(&err)

## Flags

Each analyzer can be enabled individually using its flag:

```bash
./check-all -deferloop -errs ./...
```

To enable all analyzers, use all flags:

```bash
./check-all -callsize -deferloop -errs -monkitunused ./...
```

For the callsize analyzer, you can configure thresholds:

- `-max-args int`: Maximum allowed argument size in bytes (default 64)
- `-max-results int`: Maximum allowed results size in bytes (default 256)

## Examples

Check a single package with all analyzers:
```bash
./check-all -callsize -deferloop -errs -monkitunused ./mypackage
```

Check all packages in current directory with specific analyzers:
```bash
./check-all -deferloop -errs ./...
```

## Architecture

Each analyzer is defined in its own package under `check-<analyzer>/analyzer/`. The individual checker tools (`check-errs`, `check-deferloop`, etc.) use `singlechecker.Main()` to run a single analyzer, while `check-all` uses `multichecker.Main()` to run multiple analyzers in one pass.
