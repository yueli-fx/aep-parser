# Public API Tests

`internal/aep_test` tests the public API facade and AE-validated behavior from
the outside of the package boundary.

## Contents

- Plain `*_test.go` files cover public constructors, setters, parsing, writing,
  and semantic model behavior.
- `*_shipgate_test.go` files cover AE-authored or AE-opened verification paths.
- `testutil_*` files hold shared helpers for fixture loading, host gates, and
  focused assertions.

## What Belongs Here

- Tests that should exercise the package the way a public caller would.
- Ship-gate tests that prove a shipped behavior against AE fixtures or host
  behavior.
- Regression tests for public API promises, serialization semantics, and fixture
  compatibility.

## What Does Not Belong Here

- Serializer internals tests that do not need the public API boundary. Put those
  near `internal/serializer`.
- Fixture source scripts. Use `test_data/generators/`.
- Generated local outputs. Use `test_data/generated/` or `tmp/`.
