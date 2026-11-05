# ms-test

In-memory filesystem take-home for Material Security, implemented in Go.

## Requirements

- Go 1.27+
- Optional: `golangci-lint`, Docker

## Make

| Target | Description |
|--------|-------------|
| `make test` | Tests with `-race` and `-shuffle=on` |
| `make test-fs` | Tests `internal/fs` + coverage gate (≥78%) |
| `make build` | Binary in `bin/msfs` |
| `make repl` | Interactive REPL (`go run`) |
| `make script` | Run `docs/scripts/pdf-happy-path.txt` |
| `make lint` | `golangci-lint run` |
| `make help` | All targets |

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

In the REPL, `<command> --help` (or `-h`) prints usage without changing the filesystem. `help` lists commands; `help mkdir` is the same as `mkdir --help`. The binary accepts `msfs --help` before starting the REPL.

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
