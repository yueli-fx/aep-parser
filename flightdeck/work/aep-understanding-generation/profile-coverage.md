# Profile Coverage Matrix

This matrix decides what enters Phase 1 `Profile` and what stays in the
detailed `WriteJSON` export.

| Field group | aepdissect-json | WriteJSON | Profile core decision | Evidence floor | Notes |
| --- | --- | --- | --- | --- | --- |
| Project identity | `project` path | absent as explicit path | Admit `meta.path`, `meta.schema_version`, `meta.parse_warnings` | L1_parsed | Add schema version in Phase 1. |
| Project item graph | precomp/footage source labels only | `compositions`, `footage`, `folders` with IDs | Admit normalized item table with IDs, names, type, parent/folder where available | L1_parsed | Needed for source refs and generation inventory. |
| Effect usage fingerprint | `effectUsage`, `thirdParty` | effects nested per layer | Admit summarized effect usage and plugin dependency list | L1_parsed | Keep detailed params per path under effects. |
| Composition settings | ID/name/size/fps/duration | plus tick rate, renderer, work area, bg, shutter/motion blur, markers/guides/EG | Admit ID/name/size/fps/duration/tick/render/work area/motion blur and comp markers; detail-only guides/EG until diff needs them | L1_parsed | Renderer/work area affect render and slice selection. |
| Layer identity | index/name/type | index/name/type/ID/source ID/parent ID | Admit ID, index, name, type, source ref, parent ref, matte ref | L1_parsed | Stable path uses ID where present plus index/occurrence fallback. |
| Layer timing | in/out points | start/duration/stretch | Admit normalized in/out/start/duration/stretch | L1_parsed | Required for active intervals and render sentinel selection. |
| Layer flags | visible/blend only | broad flags and switches | Admit visible, blend, 3D, solo, shy, locked, adjustment/null/guide, motion blur, effects enabled, audio, frame blend, collapse | L1_parsed | Flags frequently explain render deltas. |
| Source refs | precomp/footage display label | raw source ID plus footage table | Admit source ID and resolved display label | L1_parsed | Display label is annotation, not identity. |
| Effects | match name/class/name/tuned/params | match name/name/parameters | Admit match name, display name, occurrence, dependency class, params, tuned params, unknown params | L1_parsed | `tuned_params` remains hint, not proof of semantic importance. |
| Effect params | static value/default/changed/expression | full property shape through parameters | Admit static value, expression, keyframes, interpolation/ease when present | L1_parsed | Defaults from dict are advisory metadata. |
| Transform/property tree | animated summary only | full `properties` list | Admit stable property paths, static values, keyframes, interpolation/ease, expression | L1_parsed | Summaries can remain fingerprint fields. |
| Keyframes | count/motion/start/end | values/interp/ease/tangents | Admit full keyframe timeline for admitted properties | L1_parsed | Required before diff and sentinel frames. |
| Expressions | expression refs from effect params | expression + enabled state per property | Admit expression text, enabled state, and parsed refs where available | L1_parsed | Expression reference extraction is heuristic unless parser-backed. |
| Masks | absent | masks, vertices, path keyframes, interp/ease | Admit mask metadata and path timeline; mark geometry evidence carefully | L1_parsed | Existing knowledge says some mask geometry may be partial. |
| Shapes | absent | shape paths and primitives | Admit shape paths/primitives for shape-layer diff | L1_parsed | Avoid full graph overreach until path identity is stable. |
| Text | absent | text source, fonts, runs, paragraphs, justification | Admit text source summary and style runs | L1_parsed | Font availability becomes render oracle metadata later. |
| Markers/guides/EG | absent | present | Admit comp/layer marker payloads with stable paths; keep guides/EG detail-only until fixture proves clone/diff need | L1_parsed | Marker fixtures now prove profile value without changing core identity model. |
| Render queue | absent | present | Detail-only for Phase 1; profile fingerprint may include presence/count | L1_parsed | Generation/replication first targets comps, not render queue. |
| Unknown/raw escapes | absent | limited by scene model | Admit explicit `unknowns` records where parser exposes uncertainty | L0_raw/L1_parsed | Do not silently omit unsupported structures. |

## Phase 1 Field Admission List

Admit immediately:

- `schema_version`
- `meta.path`
- `meta.parse_warnings`
- project item table: comps, footage, folders
- comp ID/name/size/fps/duration/tick rate/renderer/work area/motion blur
  settings
- layer ID/index/name/type/source ref/timing/stretch/core flags/blend/parent/
  matte
- effect match name/display name/dependency class/occurrence/params/tuned
  params/unknown params
- property match name/path/static value/keyframes/interpolation/ease/spatial
  tangents/expression
- masks with metadata/path timeline
- shapes with path/primitive summaries
- text source and style run summaries
- comp/layer marker payloads
- evidence records and unknown records

Keep detail-only for Phase 1:

- render queue full settings
- guides
- essential graphics controllers
- detailed output module format options

## Admission Rationale

The admitted fields are the minimum set that lets Phase 2 compare structure and
semantics by stable path, and lets Phase 3 choose sentinel frames from actual
timing/keyframe data. The detail-only fields are useful exports but do not need
to become profile core until a diff, render, or generation workflow proves they
affect the decision surface.
