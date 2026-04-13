# vprf

`vprf` is a small Go CLI that wraps `samply` for recording and exposes focused subcommands for querying Firefox Profiler JSON profiles.

On current `samply` builds, `--unstable-presymbolicate` writes a `.syms.json` sidecar next to the profile. `vprf` loads that sidecar automatically so function names resolve cleanly instead of showing raw addresses.

## Install

1. Install `samply` from its release page or installer script.
2. Build this tool:

```bash
go build ./...
```

## Record

```bash
./vprf record --output profile.json.gz --duration 10 -- go run ./tests/testdata/burn.go 12
```

## Query

```bash
./vprf summary -p profile.json.gz
./vprf top -p profile.json.gz --limit 15
./vprf diff -a baseline.json.gz -b comparison.json.gz
./vprf callers -p profile.json.gz --fn innerLoop
./vprf callees -p profile.json.gz --fn outer
./vprf threads -p profile.json.gz
./vprf hotpath -p profile.json.gz
./vprf collapsed -p profile.json.gz
```

Each command also supports `--format json`.

Commands that accept `--fn` treat it as a regular expression pattern, for example `--fn "runtime\\."` or `--fn "^inner"`.

`diff` also supports `-a`, `-b`, `--thread-a`, and `--thread-b` to compare two profiles and optionally scope each side to a different thread filter.
