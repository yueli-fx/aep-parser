# Test Data Layout

`test_data/` is for test fixtures and fixture generators.

- `fixtures/` contains stable files used directly by tests. See
  `fixtures/README.md` for fixture retention rules.
- `generators/` contains JSX scripts that author or verify fixtures in After
  Effects. See `generators/README.md` before deleting older or version-named
  scripts.
- `generated/` contains ignored local outputs from regeneration, ship-gates, or
  exploratory AE runs.

Do not add new files to the `test_data/` root. Put stable fixtures in
`fixtures/`, JSX in `generators/`, and disposable outputs in `generated/`.
