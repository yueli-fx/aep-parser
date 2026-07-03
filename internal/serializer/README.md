# Serializer Package

`internal/serializer` is the byte-to-scene and scene-to-byte translation layer
for AEP project data.

## Contents

- `parse_*.go` reads RIFX/chunk structures into `internal/scene` models.
- `lower_*.go` converts scene state back into serializable chunk payloads.
- `write_*.go` writes chunk structures while preserving required layout.
- `back_*.go` hydrates parsed chunk state back onto scene objects.
- `mutate_*.go` implements mutation paths used by public setters and builders.

## What Belongs Here

- AEP chunk layout knowledge that is needed to parse, preserve, or write bytes.
- Serialization logic that depends on both raw chunk shape and scene models.
- Tests for byte layout, lower/write behavior, and serializer-only invariants.

## What Does Not Belong Here

- Public API facade types or doc comments. Use `internal/aep`.
- Pure runtime scene methods with no byte-layout dependency. Use
  `internal/scene`.
- AE host orchestration, fixture generation, or external process automation.
