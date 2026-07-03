# Recipe Examples

`examples/recipes/` contains minimal recipe JSON inputs for public recipe
generation and validation.

## What Belongs Here

- Small examples that compile through `go run ./cmd/aeprecipe ...`.
- One-capability or one-behavior recipes that are useful in documentation,
  smoke checks, or profile verification.
- Expected-public examples that a reader can understand without private context.

## What Does Not Belong Here

- Generated `.aep` files or validation output.
- Large demonstrations that depend on private sample corpora.
- Draft recipe experiments that are still being investigated.

When adding a recipe, keep the filename descriptive and prefer a narrow example
over a broad kitchen-sink file.
