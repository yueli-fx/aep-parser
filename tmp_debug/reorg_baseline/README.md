baseline commit: 3f98e136cda75945c88d8bba880928f8915ed2a5

## reorg_baseline

Round-trip byte baseline for the `internal/aep` package reorg.

### Generating the baseline

From the repo root:

```
go run ./tmp_debug/reorg_baseline
```

This walks `test_data/` (and `data/` if present), opens each `.aep` fixture via
`aep.Open`, writes it back via `proj.WriteAEP`, sha256s the output, and writes
the result to `tmp_debug/reorg_baseline/baseline.json`.

### Consuming the baseline

`internal/aep/reorg_baseline_test.go` → `TestReorgRoundtripBaseline` reads
`baseline.json` and re-runs the same Open→WriteAEP→sha256 for every fixture.
If `baseline.json` is absent the test skips (safe on main, guards only during
the reorg window).

Run with:

```
go test -count=1 ./internal/aep/ -run TestReorgRoundtripBaseline -v
```

### Expiry

Delete `tmp_debug/reorg_baseline/` and `internal/aep/reorg_baseline_test.go`
once the package reorg lands and the file-rename commits are verified
byte-identical.
