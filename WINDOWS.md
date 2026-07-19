# Windows: patching samply

`vprf record` shells out to [`samply`](https://github.com/mstange/samply).
On Windows, the published `samply` 0.13.1 (and its bundled `pdb-addr2line` 0.11.x)
panics during symbolication on PDBs that contain a `Char8` type:

```
thread 'main' panicked at
  pdb-addr2line-0.11.2\src\type_formatter.rs:1057:18:
  Unknown PrimitiveKind Char8 in emit_primitive
```

Upstream fixed this in `pdb-addr2line` 0.12.0, but `samply`'s `Cargo.lock` pins
0.11.x, so no release of `samply` today contains the fix. The same single line
must be backported into the vendored copy you build.

## Build a working `samply` on Windows

```bash
git clone --depth 1 https://github.com/mstange/samply.git
cd samply
git fetch --tags --depth 1 origin tag samply-v0.13.1
git checkout samply-v0.13.1
```

Vendor and patch `pdb-addr2line` 0.11.2:

```bash
mkdir -p vendor/pdb-addr2line
cd vendor/pdb-addr2line
tar xzf "$(cygpath -u "$USERPROFILE")/.cargo/registry/cache/index.crates.io-1949cf8c6b5b557f/pdb-addr2line-0.11.2.crate"
cd ../..
```

Apply this one-line fix to `vendor/pdb-addr2line/pdb-addr2line-0.11.2/src/type_formatter.rs`
(line ~1057):

```diff
             PrimitiveKind::Bool32 => "bool32_t",
             PrimitiveKind::Bool64 => "bool64_t",
             PrimitiveKind::HRESULT => "HRESULT",
+            PrimitiveKind::Char8 => "char8_t",
             _ => panic!("Unknown PrimitiveKind {:?} in emit_primitive", prim.kind),
         };
```

Wire the vendored copy into samply's build by appending to `Cargo.toml`:

```toml
[patch.crates-io]
pdb-addr2line = { path = "vendor/pdb-addr2line/pdb-addr2line-0.11.2" }
```

Resolve the new lock entry and build:

```bash
cargo update -p pdb-addr2line --precise 0.11.2
cargo build --release --bin samply
```

Replace the system `samply.exe` with the patched build (or `cargo install --path .`):

```bash
cargo uninstall samply
cp target/release/samply.exe "$USERPROFILE/.cargo/bin/samply.exe"
```

## Verification

```bash
vprf record -o profile.json.gz -- ./your-binary.exe
vprf top -p profile.json.gz --limit 5   # symbols, not raw addresses
```

If `top` shows raw addresses after a clean capture, the `vprf` sidecar loader
also needs the matching fix in `profile/query.go::normalizeLibName` (which
strips the same `.dll/.exe/.pdb/.sys/.ocx` suffixes).
