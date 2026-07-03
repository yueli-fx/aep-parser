# Internal Watch Review

Follow-up on `internal-ledger.md` large-package watch areas.

## Summary

No package should be split inside the hygiene survey. The watch areas are real, but each needs a dedicated plan because package boundaries are coupled to public facade docs, embedded templates, or ship-gate conventions.

## Findings

| Package | Current shape | Recommendation |
|---|---|---|
| `internal/serializer` | 78 prod Go, 17 tests, 306 embedded template/data files. Natural clusters are `parse` (20), `lower` (14), `mutate` (35), and `back` (10). | Keep as one package for now. If split later, start with a design for `serializer/templates` ownership and cycle risk between parse/lower/mutate helpers. |
| `internal/scene` | 50 prod Go, 3 tests. Mostly `scene_*` typed model and write interfaces. | Do not split now. It is the runtime model behind the public facade; premature subpackages would likely add import noise without reducing risk. |
| `internal/aepmigrate` | 105 prod Go, 45 tests plus one JSON. Heavy surface is convert/matrix/versioned migration logic. | Candidate for a dedicated migration-maintenance plan, not a hygiene cleanup. Split only if matrix/report/verification responsibilities can be separated without weakening command workflows. |
| `internal/aep_test` | 258 black-box tests. `testutil*` is small; most files are one capability/gate per test file. | Keep as an integration/ship-gate suite. Cleanup should focus on naming, stale fixtures, and generator ownership, not moving tests into many packages. |

## Next Actions

1. Treat `serializer` and `aepmigrate` as future refactor candidates only after current fixture/generator provenance is clean.
2. Keep `aep_test` dense but auditable by improving generator/fixture manifests rather than splitting the package.
3. Do not add package splits to this survey's cleanup batch.
