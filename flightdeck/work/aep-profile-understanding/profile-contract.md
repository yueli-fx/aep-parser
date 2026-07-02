# Profile Contract

This contract is the Phase 1 entry gate for extracting `internal/profile`.

## Evidence Levels

| Level | Meaning | Use in Phase 1 |
| --- | --- | --- |
| `L0_raw` | Bytes/chunks exist but are not structurally decoded. | Unknown/raw records only. |
| `L1_parsed` | Parser exposes deterministic structured fields. | Default for parsed profile fields. |
| `L2_roundtrip` | Read/write/read preserves the field. | Capability mapping, not required for all read fields. |
| `L3_ae_accept` | AE opens or renders the written construct. | Writer capability confidence. |
| `L4_render` | Selected rendered pixels match tolerance. | Render oracle output. |
| `L5_user_verified` | User/human accepts target result. | Showcase acceptance and final promotion. |

Evidence is field-scoped. A layer can be parsed at `L1_parsed` while its render
fidelity remains unproven, and a writer API can have `L3_ae_accept` while a
specific visual clone still fails `L4_render`.

## Stable Path Object

Profile records that can be diffed must carry:

- `path`: machine path using IDs, indexes, match names, and occurrence indexes.
- `display_path`: human-readable path with names.
- `identity`: structured identity fields.
- `evidence`: strongest evidence for this record.

Example:

```yaml
path: comps.by_id[17].layers.by_id[42].effects.by_match_name["ADBE Fill"]#0.params.by_match_name["ADBE Fill-0002"]
display_path: comps["main"].layers[3:"white flash"].effects[0:"Fill"].params["Color"]
identity:
  comp_id: 17
  layer_id: 42
  layer_index: 3
  effect_match_name: ADBE Fill
  effect_occurrence: 0
  property_match_name: ADBE Fill-0002
evidence:
  level: L1_parsed
  source: internal/scene
  confidence: high
```

## Path Rules

- Use comp IDs when present; use `comps.by_index[n]` only for ID-less records.
- Use layer IDs when present; keep layer index as annotation and fallback.
- Use effect `matchName` plus zero-based occurrence, because duplicate effects
  are legal.
- Use property `matchName` plus occurrence or property-group lineage for
  duplicate properties.
- Keep display names out of machine identity unless there is no other identity
  source.
- Emit `unstable_path` rather than creating a text-only path when identity is
  insufficient.
- Ignore rules must target `path`, `kind`, and optional `condition`, never only
  display text.

## Unknowns

Unknown records are part of the profile contract, not errors to hide:

```yaml
unknown:
  path:
  reason: missing_parser | raw_only | unstable_identity | unsupported_type | environment_dependent
  evidence:
    level: L0_raw
    source:
    confidence:
```

Phase 1 should emit unknowns where the existing parser can identify that a
construct exists but cannot surface it as stable structured data.

## Diff Readiness

The Phase 1 profile is diff-ready only when these records have stable paths:

- compositions
- project items
- layers
- source refs
- parent/matte refs
- effects
- effect parameters
- transform/properties
- keyframes
- masks
- shapes
- text source/style summaries

If a record cannot satisfy the path rules, it should carry `unstable_path` and
be visible to later gap reporting.

## Phase 1 Entry Gate

Human review is required before extracting `internal/profile`.

Phase 1 may start only when reviewers accept:

- `phase0-audit.md`
- `profile-coverage.md`
- this profile contract
- the decision that `Profile` is a normalized layer, while `WriteJSON` remains
  the detailed export
- the decision that Phase 1 includes stable paths and evidence records before
  diff tooling
