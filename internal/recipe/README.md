# Recipe Package

`internal/recipe` validates and compiles JSON recipe inputs into AEP project
state through the public facade.

## What Belongs Here

- Recipe schema types, validation rules, and compiler code.
- Materialization logic that turns recipe JSON into `internal/aep` project
  construction calls.
- Capability checks that describe what recipe generation can express.

## What Does Not Belong Here

- Public example recipe files. Use `examples/recipes/`.
- Generated recipe documentation. Use `internal/recipedoc` and `docs/`.
- Low-level serializer byte layout code. Use `internal/serializer`.

Keep recipes deterministic and validation-first so examples can double as
repeatable verification inputs.
