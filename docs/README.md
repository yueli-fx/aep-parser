# Docs

`docs/` contains the generated public documentation surface for the repository.
It is for reader-facing API, recipe, and capability references, not private
planning notes.

## Start here / 阅读入口

- [中文 README](../README.md) / [English README](../README.en.md): installation, public SDK, CLI, compatibility, and knowledge navigation.
- [SDK 与集成指南](usage.md) / [SDK and integration guide](usage.en.md): complete Go examples, HTTP uploads, response handling, and compatibility.
- [ProjectJSON v2](project-json-v2.md): property snapshot contract.
- [Recipe reference](recipe.md) and [examples](../examples/recipes/README.md): supported generation inputs.
- [Capability matrix](capabilities.md): field-level scope and verification boundaries.
- [Self-hosted reports](self-hosted-reports.md): corpus analysis and monitoring (Chinese).
- [Open-source audit](open-source-audit.md): review scope, results, and pending release checks (Chinese).

The generated type references describe the internal `internal/aep` facade as well as its model. They are not a promise that every listed symbol is exported by the root Go SDK; external consumers should follow the root README and the versioned Recipe/ProjectJSON contracts.

- [Format and engineering knowledge](knowledge/README.md): reverse-engineering notes, implementation decisions, and verification workflows.
- [Runnable SDK examples](../examples/sdk/README.md) and [composed projects](../examples/projects/README.md).

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

- Draft specs, active plans, or work logs. Keep these in the maintainer’s work desk, outside the public documentation.
- Reverse-engineering notes. Use `docs/knowledge/<domain>/` after they are
  durable.
- Hand edits to generated API/capability/recipe files unless the generator is
  also updated.
