# Scene Package

`internal/scene` is the runtime project model shared by the parser, serializer,
public API facade, recipes, and migration code.

## What Belongs Here

- In-memory project, composition, layer, property, mask, shape, text, footage,
  and render-queue models.
- Runtime methods that describe scene semantics without depending on raw chunk
  byte layout.
- Interfaces and contracts consumed by serializer/write paths.

## What Does Not Belong Here

- Raw RIFX/chunk parsing or byte layout rules. Use `internal/serializer`,
  `internal/rifx`, or `internal/codec`.
- Public facade docs or exported user-facing aliases. Use `internal/aep`.
- Host-assisted verification, AE automation, or fixture generation.

Keep this package focused on semantic state. If a change needs to know exact
chunk offsets, it probably belongs in the serializer layer instead.
