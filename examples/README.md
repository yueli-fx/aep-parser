# Examples

`examples/` contains small, deterministic inputs that demonstrate public
project-generation behavior.

## Contents

- `recipes/` contains JSON recipe examples used by recipe documentation,
  validation, and profile checks.
- `integrations/` contains small public workflow examples for placing the
  headless toolkit in existing pipelines.

## What Belongs Here

- Minimal examples that are readable, reproducible, and useful as public usage
  references.
- Recipe inputs that can be compiled and validated by the normal recipe tools.
- Small examples that explain one behavior or capability at a time.

## What Does Not Belong Here

- Generated AEP outputs, renders, or local verification artifacts. Use `tmp/` or
  ignored generated output locations.
- Large sample projects. Use the ignored `data/samples/` corpus.
- Private reverse-engineering experiments that are not useful as examples.
