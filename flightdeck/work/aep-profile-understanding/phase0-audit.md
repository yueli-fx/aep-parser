# Phase 0 Export Surface Audit

## Purpose

Phase 0 decides what the stable profile is allowed to be before extracting
`internal/profile`. The audit compares the current compact diagnostic export
(`cmd/aepdissect -json`) with the broad parsed-scene export
(`Project.WriteJSON`) and records which fields should enter the first profile
contract.

## Export Surfaces

- `cmd/aepdissect -json` is a diagnostic/understanding profile embedded inside
  the CLI. It exposes a compact project summary: project path, effect usage,
  third-party effect list, comps, layers, effect params, tuned params, animated
  transform/effect summaries, expression references, and precomp/footage source
  labels.
- `Project.WriteJSON` is a broad parsed-scene export. It exposes project items,
  render queue, comp settings, layer identity/flags/timing/source IDs,
  properties/keyframes/ease, effects, masks, shape paths/primitives, text
  source/style runs, markers, guides, folders, and footage.
- Phase 1 must not copy either surface wholesale. `Profile` should be a
  normalized comparison/learning layer with evidence and stable paths;
  `WriteJSON` remains the detailed export.

## Fixture Set

- `flightdeck/showcase/booyah-clone/booyah-clone.aep`: current real-project
  stress fixture; useful for nested comps, effects, motion, and authored visual
  complexity.
- `flightdeck/showcase/text/text.aep`: focused text-source/style coverage.
- `flightdeck/showcase/masks/masks.aep`: focused mask/path coverage.
- `flightdeck/showcase/keyframes-ease/keyframes_ease.aep`: focused
  keyframe/interpolation/ease coverage.
- `flightdeck/showcase/effects/effects.aep`: focused effect/parameter coverage.
- `internal/serializer/templates/project/2025.aep`: tiny skeleton project for
  baseline item/project shape.

## Commands Run

```powershell
go run ./cmd/aepdissect -json flightdeck/showcase/text/text.aep
go run ./cmd/aepdissect -json flightdeck/showcase/masks/masks.aep
go run ./cmd/aepdissect -json flightdeck/showcase/keyframes-ease/keyframes_ease.aep
go run ./cmd/aepdissect -json flightdeck/showcase/effects/effects.aep
go run ./cmd/aepdissect -json flightdeck/showcase/booyah-clone/booyah-clone.aep
```

All five commands exited 0 and emitted JSON with `project`, `effectUsage`, and
`comps`.

Observed examples:

- `text.aep`: one comp, text layer and background layer, no text-source/style
  data in `aepdissect -json`.
- `masks.aep`: one comp, five shape layers, no mask geometry/path data in
  `aepdissect -json`.
- `keyframes_ease.aep`: one comp, animated summaries for Position with keyframe
  count, motion label, start, and end; no full keyframe values/ease arrays.
- `effects.aep`: effect usage histogram and per-layer effect params with
  `changed` and `tuned` hints.
- `booyah-clone.aep`: effect usage and many animated summaries across nested
  comps; useful as stress input, but output is too shallow for stable diff.

## Observed aepdissect-json Shape

`aepdissect -json` is compact and comparison-friendly but shallow:

- Has project path, effect usage histogram, third-party effect list, comp
  ID/name/size/fps/duration, layer index/name/type/blend/visible/timing/source
  label, parent/matte index, effects, effect params, changed/tuned hints,
  animated summaries, and expression refs.
- Does not expose project folders, full footage metadata, render queue, comp
  renderer/work area/motion blur settings, layer IDs, many layer flags, raw
  source IDs, full property trees, full keyframe values/ease arrays, masks,
  shapes, text source/style runs, markers, guides, essential graphics, or
  unknown/raw escape hatches.
- Uses human display labels in several places and parent/matte by layer index,
  so it is not yet a stable diff path contract.

## Observed WriteJSON Shape

`Project.WriteJSON` is broad and deterministic, but it is a detailed parsed
export rather than a curated profile:

- Project-level: compositions, footage, folders, render queue.
- Composition-level: ID, name, dimensions, frame rate, duration, tick rate,
  background color, resolution factor, renderer, work area, motion blur/shutter
  settings, layers, markers, guides, motion graphics template name, essential
  graphics controllers.
- Layer-level: index, name, type, ID, parent ID/name, source ID, timing,
  stretch, quality, label, blending mode, track matte, comments, 3D/solo/shy/
  locked/visible/adjustment/null/guide/motion blur/effects/audio/frame blend/
  collapse/shape flags.
- Content-level: properties, effects and their parameters, markers, masks,
  shape paths, shape primitives, text source/style runs/paragraphs.
- Property-level: name, match name, static value, expression, expression enabled
  state, keyframes, interpolation, temporal ease, and spatial tangents where
  surfaced.
- Render-level: queue items, output modules, render settings, output module
  settings, and format options.

## WriteJSON Limits for Profile Use

- It is one-way export; there is no corresponding `ReadJSON`.
- It is intentionally broad, so direct diffs would be noisy.
- It lacks evidence levels, schema version, stable path objects, field admission
  metadata, and explicit unknown classifications.
- Some values are display-friendly rather than profile contracts, such as names
  beside IDs.
- Large projects can produce bulky output; Phase 1 profile should summarize
  repeated/noisy fields unless detailed values are needed for diff or
  generation.

## Phase 0 Conclusion

The current project already has both ingredients needed for a profile, but not a
profile contract:

- `aepdissect -json` should supply the compact fingerprint and technique-facing
  fields.
- `WriteJSON` should supply admitted detail fields for identity, flags,
  keyframes, masks, shapes, text, and source graph.
- `Profile` should be a new normalized layer that adds schema version, evidence,
  stable paths, explicit unknowns, and field admission boundaries.

Next human review point: approve the Phase 0 artifacts before extracting
`internal/profile`.
