# ms-test — Claude Code

In-memory filesystem take-home (Material Security). Module `github.com/perylemke/ms-test`.

## Workflow

- Branch `develop`; one-line conventional commits in English per slice.
- TDD: tests before implementation; `go test -race -shuffle=on ./...` before commit.
- Code and README in English.

## Layout

- `internal/fs` — library (Session, FS store, inodes, extensions).
- `cmd/msfs` — REPL CLI.

## Go

- Sentinel errors with `errors.Is`; wrap with `%w`.
- No panic in normal paths.
- List directory children in creation order.

## Testing

- Table-driven subtests; `fs_test` package for public API.
- Cover PDF script and error sentinels per slice.

See `.cursor/rules/` for the same conventions in Cursor.
