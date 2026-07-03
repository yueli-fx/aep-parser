# Recipe Documentation Package

`internal/recipedoc` builds generated documentation and indexes for the recipe
schema and example set.

## What Belongs Here

- Reflection and rendering logic for recipe schema docs.
- Builders for recipe indexes, summaries, validation metadata, and JSON/markdown
  outputs.
- Tests that lock generated recipe documentation structure.

## What Does Not Belong Here

- Recipe compiler behavior. Use `internal/recipe`.
- Hand-authored public docs. Generated outputs live under `docs/`.
- Example recipe inputs. Use `examples/recipes/`.

When generated docs drift, fix the model or renderer here before editing output
files.
