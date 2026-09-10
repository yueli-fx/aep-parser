# JSX generator cleanup provenance

JSX generator cleanup requires checking fixture manifest entries, produced fixture names, tracked generated outputs, and baseline/evidence references before deleting a zero-reference script.

A `.jsx` generator can have zero direct references by filename and still be active provenance. The common cases:

- It is listed in `scripts/fixtures/fixtures_manifest.json`, so `regen_fixtures.ps1` owns it even if Go tests never mention the script path.
- Tests or knowledge reference the produced `.aep`, not the generator filename.
- The generator path is stale but the output fixture is still a truth source; fix the path/manifest before deleting anything.
- A generated ship-gate `.aep` may only appear in `tools/debug/split_roundtrip_baseline/baseline.txt` or historical knowledge. Decide whether that evidence is still useful before removing script + fixture + hash together.

Cleanup order:

1. Extract output names from the script (`.aep`, `.done`, `.json`, `.txt`, `.png`).
2. Search those output names across `internal`, `docs/knowledge`, `registry`, `docs`, `scripts`, `tools`, and `test_data`.
3. Check `scripts/fixtures/fixtures_manifest.json` before treating a zero-reference script as unused.
4. Check `registry/capability_atoms.json` for required glob atoms before deleting the last script in a prefix family; otherwise `aepregistry audit` will fail on a missing dependency.
5. If the output is tracked but unreferenced, delete or archive the script and output in one commit; also remove stale baseline hashes.
6. If the output is still a fixture/template truth source, repair provenance instead of deleting the generator.
