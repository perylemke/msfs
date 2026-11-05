# Manual test script

Run commands via `go run ./cmd/msfs` (REPL) or place lines in a file and use `go run ./cmd/msfs run file.txt`.

Expected errors are shown as `error: ...` on stderr.

Use `<command> --help` (or `help <command>`) to see usage; help does not modify the filesystem.

## PDF happy path

| Step | Command | Expected stdout |
|------|---------|-----------------|
| 1 | `mkdir school` | (empty) |
| 2 | `cd school` | (empty) |
| 3 | `pwd` | `/school` |
| 4 | `mkdir homework` | (empty) |
| 5 | `cd homework` | (empty) |
| 6 | `mkdir math` | (empty) |
| 7 | `mkdir lunch` | (empty) |
| 8 | `mkdir history` | (empty) |
| 9 | `mkdir spanish` | (empty) |
| 10 | `rmdir lunch` | (empty) |
| 11 | `ls` | `math history spanish` |
| 12 | `pwd` | `/school/homework` |
| 13 | `cd ..` | (empty) |
| 14 | `mkdir cheatsheet` | (empty) |
| 15 | `ls` | `homework cheatsheet` |
| 16 | `rmdir cheatsheet` | (empty) |
| 17 | `cd ..` | (empty) |
| 18 | `pwd` | `/` |

Script: [scripts/pdf-happy-path.txt](scripts/pdf-happy-path.txt)

## Core error cases

| Scenario | Commands | Expected error |
|----------|----------|----------------|
| cd missing | `cd nope` | `not found` |
| duplicate mkdir | `mkdir a` then `mkdir a` | `already exists` |
| invalid name | `mkdir bad/name` | `invalid name` |
| rmdir not empty | `mkdir d`, `cd d`, `mkdir c`, `cd ..`, `rmdir d` | `directory not empty` |
| rmdir not child | `mkdir a`, `mkdir b`, `cd a`, `rmdir b` | `not a child of working directory` |
| cd .. at root | `cd ..` then `pwd` | `/` |
| read missing file | `cat missing` | `not found` |
| read directory | `mkdir d`, `cat d` | `is a directory` |
| write directory | `mkdir d`, `write d x` | `is a directory` |
| move collision | `touch a`, `touch b`, `mv a b` | `already exists` |
| rmdir file | `touch f`, `rmdir f` | `not a directory` |

## Paths extension

| Scenario | Commands | Expected |
|----------|----------|----------|
| mkdir -p | `mkdir-p /a/b/c`, `cd /a/b`, `pwd` | `/a/b` |
| relative .. | `cd ..`, `pwd` from `/a/b` | `/a` |
| read/write at path | `mkdir-p /x/y`, `touch /x/y/f`, `write /x/y/f hi`, `cat /x/y/f` | `hi` |
| cd through file | `touch f`, `cd f` | `not a directory` |

## Move / copy extension

| Scenario | Commands | Expected |
|----------|----------|----------|
| directory merge | `mkdir-p /src/a`, `touch /src/a/one`, `write /src/a/one 1`, `mkdir-p /dst/a`, `touch /dst/a/two`, `mv /src/a /dst/a`, `cat /dst/a/one` | `1` |
| file copy rename | `touch a`, `touch b`, `cp a b`, `cat b~1` | empty file readable |
| type mismatch | `mkdir d`, `touch f`, `mv f d` | `type mismatch` |

## Walk / regex extension

| Scenario | Commands | Expected |
|----------|----------|----------|
| findre | `mkdir-p /d`, `touch /d/foo-1.txt`, `findre 'foo-\d+\.txt'` | `/d/foo-1.txt` |

## Symlink / hardlink extension

| Scenario | Commands | Expected |
|----------|----------|----------|
| symlink read | `touch /f`, `write /f x`, `ln-s /f /link`, `cat /link` | `x` |
| hardlink + mv | `touch orig`, `write orig z`, `ln orig alias`, `mv orig moved`, `cat alias` | `z` |

## Permissions extension

| Scenario | Commands | Expected |
|----------|----------|----------|
| deny write | `user add other`, `touch secret`, `chmod secret ro`, `su other`, `write secret no` | `permission denied` |

## Stream extension

| Scenario | Commands | Expected |
|----------|----------|----------|
| second writer | `touch f`, open two writers via API tests | `writer already open` (see unit tests) |

Stream seek-after-move is covered by `go test -run TestStreamSeekAfterMove`.

## Notifications / shared FS

| Scenario | How | Expected |
|----------|-----|----------|
| cancel rmdir | Register before hook returning cancel; `rmdir d` | `operation canceled` (unit test) |
| two sessions | Session A creates file; session B `cat` same path | success (unit test) |
