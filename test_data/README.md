# Test Data Layout

`test_data/` is for test fixtures and fixture generators.

- `fixtures/` contains stable files used directly by tests.
- `generators/` contains JSX scripts that author or verify fixtures in After
  Effects.
- `generated/` contains ignored local outputs from regeneration, ship-gates, or
  exploratory AE runs.

Do not add new files to the `test_data/` root. Put stable fixtures in
`fixtures/`, JSX in `generators/`, and disposable outputs in `generated/`.
