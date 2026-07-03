# Folder usage policy for generated evidence
SUMMARY: `tmp/` is disposable; durable generated evidence and long-lived registry truth must be promoted to owned tracked locations before it becomes a dependency.
READ WHEN: adding generated artifacts, referencing generated files from registry/current/coverage JSON, cleaning tmp/data/examples/test_data, or deciding where batch evidence should live

---

Use directories by lifecycle, not by convenience:

- `tmp/`: disposable reports, scratch outputs, one-off run logs, and rebuild caches. A clean clone must not need any file from `tmp/` to explain project state or pass registry gates. Do not point durable registry/current/coverage/evidence state at `tmp/`.
- `registry/`: machine-readable truth sources and governance ledgers. State that the tools read to decide status belongs here.
- `registry/evidence/<topic>/`: durable generated evidence that is required by registry state, coverage ledgers, contract gates, or future continuation. These files are generated, but they are tracked because they are proof artifacts.
- `data/samples/`: user-approved local corpus for broad sample validation. It is not a cleanup target and should not be promoted implicitly into truth ledgers.
- `examples/recipes/`: tracked atomic recipe specs used to generate deterministic evidence.
- `test_data/fixtures/` and `test_data/generators/`: tracked regression fixtures and their generators.
- `test_data/generated/`: rebuildable generated test data; keep it out of truth ledgers unless a file is explicitly promoted elsewhere.

Promotion rule:

If a generated file is referenced by `registry/evidence.json`, `registry/versioned_aep_migration_current.json`, `registry/versioned_aep_migration_coverage.json`, a capability atom, or a contract gate, store it under `registry/evidence/<topic>/` and update producer commands to write there by default.

Cleanup rule:

Before deleting or pruning generated outputs, run the registry cleanup/layout gates and treat `tmp/registry_*.json` as disposable reports. If a cleanup report says a `tmp` file is protected only because a long-lived JSON references it, migrate that file to `registry/evidence/<topic>/` first; do not preserve `tmp` as a hidden archive.
