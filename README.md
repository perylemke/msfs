# ms-test

In-memory filesystem take-home for Material Security, implemented in Go.

## Requirements

- Go 1.27+

## Run tests

```bash
go test -race -shuffle=on ./...
```

Coverage gate (CI): `internal/fs` must stay at or above 78%.

## Run the CLI

Interactive REPL:

```bash
go run ./cmd/msfs
```

Run a script file (one command per line, `#` comments):

```bash
go run ./cmd/msfs run docs/scripts/pdf-happy-path.txt
```

## Manual test scenarios

See [docs/manual-tests.md](docs/manual-tests.md) for step-by-step happy paths and error cases matching the PDF and extensions.

## Layout

| Path | Purpose |
|------|---------|
| `internal/fs` | Filesystem library (`Session`, shared `FS`) |
| `cmd/msfs` | Thin REPL CLI |
| `.github/workflows` | CI: test, lint, security, docker scan, release |

## Git workflow

Development happens on branch `develop`. Conventional one-line commit messages in English.
