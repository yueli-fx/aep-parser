# Fixtures

`test_data/fixtures/` contains stable fixture inputs used directly by tests.

## What Belongs Here

- Small AEPs or fixture files that are committed because tests depend on them.
- Fixture subdirectories with a clear generator, source script, or provenance
  note.
- Binary evidence that is intentionally part of regression coverage.

## What Does Not Belong Here

- Regenerated local output from a test run. Use ignored `test_data/generated/`.
- Exploratory AE output that has not become a stable fixture. Use `tmp/`.
- Large sample corpora. Use ignored `data/samples/`.

When replacing a fixture, keep or update the matching generator/provenance in
`test_data/generators/` so future cleanup can prove why the file exists.
