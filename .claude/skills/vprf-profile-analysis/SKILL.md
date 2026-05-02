---
name: vprf-profile-analysis
description: Analyze CPU profiles from samply (Firefox Profiler JSON format) using the vprf CLI. Use this skill when the user asks to investigate performance, analyze a .json or .json.gz profile, find hot functions, examine call paths, compare two profiles, or generate flamegraph input. Triggers on mentions of: samply, vprf, CPU profile, flamegraph, hot path, hot function, profile diff, or when the user provides a path to a .json.gz or .json file that looks like a profile.
---

# vprf Profile Analysis

`vprf` is a CLI tool installed at `~/.local/bin/vprf` that queries CPU profiles in the Firefox Profiler JSON format (typically produced by `samply`). It is designed for AI consumption: structured table and JSON output, regex-based filtering, zero-dependency Go binary.

## When to use this skill

- User asks to analyze a `.json.gz` or `.json` profile file
- User mentions "samply", "flamegraph", "CPU profile", "hot function", "hot path"
- User asks to compare two profiles (regression analysis, A/B perf testing)
- User asks to find what's slow, what's calling what, or which thread is hottest
- User wants flamegraph collapsed-stack output for tools like `inferno-flamegraph`

## Available commands

All query commands accept `-p <profile>` and `--format json` (except `record`, `collapsed`, and `diff` which have specific flags).

| Command | Purpose |
|---|---|
| `vprf record -o <out> --duration <s> -- <cmd>` | Wrap `samply record`. Auto-presymbolicates. |
| `vprf summary -p <profile>` | One-shot overview: duration, samples, hottest thread/function (skips idle funcs) |
| `vprf top -p <profile> [--limit N] [--sort self\|total]` | Top functions ranked by self or total samples |
| `vprf callers -p <profile> --fn <regex>` | Direct callers of matching function(s) |
| `vprf callees -p <profile> --fn <regex>` | Direct callees of matching function(s) |
| `vprf threads -p <profile>` | Per-thread sample breakdown |
| `vprf hotpath -p <profile> [--limit N]` | Most-sampled full call stacks |
| `vprf collapsed -p <profile>` | Collapsed-stack format `func1;func2;func3 count` for flamegraph tools |
| `vprf diff -a <profileA> -b <profileB> [--limit N] [--sort delta\|self_a\|self_b\|name]` | Per-function delta between two profiles |

Common flags (where applicable):
- `--thread <name-or-tid>` — substring match on thread name or exact TID match
- `--format json` — structured JSON output (preferred for parsing)
- `--fn <regex>` — RE2 regex pattern. Always treated as regex, never substring.
- `--limit N` — cap rows (default 10 for query commands, 0 = unlimited for collapsed)

## Workflow recommendations

### Investigating "what's slow?"

```bash
vprf summary -p profile.json.gz
vprf top -p profile.json.gz --limit 20
vprf threads -p profile.json.gz
```

If the top is dominated by idle/wait functions (e.g. `__psynch_cvwait`, `kevent`, `nanosleep`, `runtime.gopark`), narrow to a working thread with `--thread <name>` or filter to non-idle by using `--fn` with a negative pattern via grep on JSON output.

### Investigating a specific function

```bash
vprf callers -p profile.json.gz --fn "MyFunction"
vprf callees -p profile.json.gz --fn "MyFunction"
```

The `--fn` pattern is a regex. Use `^Foo$` for exact match, `Foo.*` for prefix, `(?i)foo` for case-insensitive.

If multiple functions match, vprf prints them to stderr and aggregates results across all matches.

### Comparing two profiles (regression hunting)

```bash
vprf diff -a baseline.json.gz -b candidate.json.gz --limit 30 --sort delta
```

Output columns: `delta_self`, `pct_chg`, `self_a`, `self_b`, `delta_total`, `pct_chg_total`, `function`, `module`. Positive `delta_self` means candidate is slower in that function.

### Generating flamegraphs

```bash
vprf collapsed -p profile.json.gz | inferno-flamegraph > flame.svg
```

Or pipe to Brendan Gregg's `flamegraph.pl`. Output is one stack per line, `;`-separated frames, space, sample count.

### Finding which thread is hot

```bash
vprf threads -p profile.json.gz --format json
```

Filter subsequent queries with `--thread <name>` (substring) or `--thread <tid>` (exact).

## Data format notes

- Profiles are typically `.json.gz` (gzip-compressed Firefox Profiler JSON).
- `samply --unstable-presymbolicate` writes a `.syms.json` sidecar next to the profile. `vprf` auto-loads it for symbol resolution.
- The `record` command always passes `--unstable-presymbolicate` so symbols are correctly resolved.
- Sample weights matter: each sample has a weight (often 1, sometimes higher). All counts are weighted.

## Language / symbol resolution support

Symbol resolution is delegated to `samply`'s presymbolicate step, which works for any language whose binary has standard symbol tables or debug info (DWARF, PDB, Mach-O nlist, ELF symtab). Confirmed working:

- **Go** — full symbol names (`main.innerLoop`, `runtime.gopark`, etc.) when built normally
- **C/C++** — fully demangled including templates, namespaces, and class methods (e.g., `mynamespace::innerLoop<long>(long)`, `Worker::doWork(long)`). Works with both debug and optimized+stripped builds, as long as the symbol table is present.
- **Rust** — demangled symbols (Rust uses similar mangling to C++ via Itanium ABI in many cases)
- **System libraries** — symbols resolve from system frameworks (libsystem, dyld, etc.) on macOS

If symbols appear as raw addresses (`0x1234abcd`), it usually means:
1. The `.syms.json` sidecar is missing (the binary was stripped of all symbols, including the symbol table)
2. The sidecar failed to load (currently silently ignored — see codebase TODO)
3. The library is JIT-compiled (e.g., V8, JVM) without a symbol map

## Output handling for AI

Always prefer `--format json` when you intend to parse, filter, or post-process. JSON output:
- `top`, `threads`, `callers`, `callees`, `hotpath`, `diff`: array of objects
- `summary`: single object
- `collapsed --format json`: array of `{stack, count}` objects (text format is `stack count\n` lines)

## Common pitfalls

- The default `top` includes idle/wait functions which often dominate. Use `summary` for the auto-filtered top non-idle function, or filter with `--fn` to a regex matching what you care about.
- `--fn` is RE2 regex. `vprf top --fn .` matches everything. `vprf top --fn "^main$"` matches only `main`. RE2 has no backreferences but supports `(?i)` for case-insensitive.
- `--thread` does case-insensitive substring matching on thread name, falling back to exact TID match. Multi-thread profiles often have anonymous threads named `Thread <tid>`.
- Profile time is in milliseconds in the JSON; `vprf summary` reports seconds.
- `diff` requires both profiles to have been collected with similar sampling rates and durations to be meaningful — `pct_chg` normalizes per-function but absolute deltas don't.

## Quick recipes

**Find the top non-idle function across all threads:**
```bash
vprf summary -p profile.json.gz
```

**Top 5 functions by total time (inclusive of callees):**
```bash
vprf top -p profile.json.gz --limit 5 --sort total
```

**Who calls anything matching a pattern:**
```bash
vprf callers -p profile.json.gz --fn "alloc|malloc"
```

**Hottest stacks in a specific thread, JSON for parsing:**
```bash
vprf hotpath -p profile.json.gz --thread "render" --format json --limit 10
```

**What got slower between two runs:**
```bash
vprf diff -a before.json.gz -b after.json.gz --sort delta --limit 20
```
