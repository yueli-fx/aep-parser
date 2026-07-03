# Docs

`docs/` contains the generated public documentation surface for the repository.
It is for reader-facing API, recipe, and capability references, not private
planning notes.

## Contents

- API markdown files such as `project.md`, `composition.md`, `layer.md`, and
  `property.md` are generated from exported public API comments.
- `capabilities.md` and `capabilities.json` are generated capability truth
  artifacts.
- `recipe_schema.json`, `recipe_index.json`, and `recipe.md` are generated from
  recipe metadata.
- `_includes/` and `superpowers/` hold supporting generated/reference material
  used by the documentation surface.

## What Belongs Here

- Generated public docs from `cmd/docgen`, `cmd/capindex`, or
  `cmd/recipedocgen`.
- Stable reader-facing documentation that describes shipped behavior.
- Machine-readable docs artifacts consumed by downstream tooling.

## What Does Not Belong Here

- Draft specs, active plans, or work logs. Use `flightdeck/work/`.
- Reverse-engineering notes. Use `flightdeck/knowledge/<domain>/` after they are
  durable.
- Hand edits to generated API/capability/recipe files unless the generator is
  also updated.
