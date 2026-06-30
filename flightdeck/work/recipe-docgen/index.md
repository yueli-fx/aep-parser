# Index — recipe-docgen

## State

This work package specifies a dedicated documentation generator for the recipe
IR in `internal/recipe`.

The recipe IR has grown enough that source files and tests are no longer the
right way to discover supported JSON fields. API docs already have
`cmd/docgen` and `cmd/capindex`, but those tools document Go API symbols and
verified write capabilities. Recipe docs need a separate contract: JSON field
paths, field types, validation rules, example usage, expected-profile support,
and the capability queries each field triggers.

## Goal

Build a generator that produces recipe-facing documentation from a stable
recipe schema contract without overloading the public API annotation system.

The first implementation should be intentionally small:

- Generate a machine-readable recipe schema from `internal/recipe` exported
  structs and `json` tags.
- Generate a human-readable field reference for authoring recipe JSON.
- Attach explicit validation notes and capability-query mappings from a
  recipe-owned registry.
- Add a drift gate so generated recipe docs cannot silently go stale.

## Non-Goals

- Do not reuse `@summary`, `@gate`, or `@verify` directly on recipe fields.
  Those annotations describe Go API symbols and AE verification boundaries.
- Do not fold recipe documentation into `docs/capabilities.md`.
  Capabilities remain the API write-surface truth source.
- Do not attempt to parse validation logic from arbitrary Go code in the first
  version. Validation behavior should be documented through an explicit
  registry.
- Do not generate a correction loop or recipe synthesizer as part of this work.

## Proposed Outputs

Generated outputs should live under `docs/` once implemented, because they are
public generated documentation:

- `docs/recipe_schema.json`
  - Machine-readable schema for tools, editors, and agents.
  - Contains field path, JSON name, Go type, required/optional state, array
    element shape, enum values where known, and capability queries where known.
- `docs/recipe.md`
  - Human-readable authoring reference.
  - Organized by recipe object: `Recipe`, `CompSpec`, `Layer`, `Transform`,
    `ShapeSpec`, `Effect`, `ExpectedProfile`, and nested specs.
  - Each field row shows path, type, required/optional state, validation note,
    capability query, and short prose.
- Optional later output: `docs/recipe_index.json`
  - A compact lookup index keyed by stable recipe field path, for agents and
    editor integrations.

The spec itself stays in `flightdeck/work/recipe-docgen/`; generated public
artifacts belong in `docs/` after implementation.

## Source Model

The generator should build one canonical field model, then render every output
from that model.

Pipeline:

1. reflect recipe structs into structural field facts
2. join recipe-owned semantic registries by stable recipe path
3. join capability details by stable recipe capability key
4. produce a canonical `FieldModel`
5. render JSON schema and markdown from the same `FieldModel`

Markdown is only a renderer. It must not have a separate data source or a
markdown-specific field registry.

The generator should use these sources of truth.

### Structural Source

Reflection over `internal/recipe` exported structs provides:

- type names
- exported fields
- JSON field names from `json` tags
- `omitempty`
- pointer-vs-value optionality
- slices and nested struct relationships

Reflection also owns type, array, object, and structural requiredness. No
registry is allowed to repeat those facts.

Reflection must ignore output/report structs unless explicitly included. The
first include set should be:

- `Recipe`
- `ProjectSpec`
- `CompSpec`
- `Layer`
- `Transform`
- `Effect`
- `EffectParam`
- `ExpressionSpec`
- all nested comp, layer, text, camera, light, shape, mask, keyframe, and
  `ExpectedProfile` specs

### Metadata Registry

Recipe-owned registries should provide only semantic notes that reflection
cannot know safely:

- short summary
- validation note
- enum values
- numeric range
- array length rules
- required-if rules
- examples or example references

Do not store these in metadata registries:

- path, except as the map key
- Go type
- JSON type
- required/optional state derived from struct shape
- array/object shape
- field ordering

Recommended registry shape:

```go
package recipedoc

type FieldMeta struct {
    Summary    string
    Validation string
    Enum       []string
    Example    string
}
```

The registry should key by stable recipe path, for example:

- `comps[].width`
- `comps[].layers[].type`
- `comps[].layers[].transform.position_keyframes[].in_ease.influence`
- `comps[].layers[].shape.gradient_fill.color_stops[].offset`
- `expected_profile.keyframes[].keyframes[].value`

This keeps documentation intent explicit and reviewable. It also avoids brittle
logic that tries to infer all constraints by parsing `ValidateWithCapabilities`.

To keep review small as the field count grows, metadata should be split by
purpose:

- `summary_registry.go`
- `validation_registry.go`
- `capability_registry.go`
- `example_registry.go`

Each registry should use the same stable recipe path keys, but own only its
specific concern.

## Stable Identifiers

Recipe paths are public stable identifiers.

Once a path appears in `docs/recipe_schema.json`, it should remain valid unless
the recipe schema version changes. Renaming a Go struct or field must not by
itself rename the recipe path. A JSON field rename is a schema change and must
be handled as a compatibility decision, not as a mechanical refactor.

Examples:

- `comps[].layers[].effects[].params[]` remains stable even if the Go type
  `EffectParam` is later renamed.
- `expected_profile.keyframes[].keyframes[].value` remains stable even if the
  backing Go type is moved to another file.

The drift gate must check both directions:

- every included recipe field has a generated schema entry
- every registry key references an existing generated recipe path

## Capability Mapping Rules

Recipe docs should reference capability queries, not duplicate capability
truth.

Recipe-facing capability references should use stable recipe capability keys,
not raw Go symbol names. A separate mapping layer resolves those keys to
capindex queries.

Example stable keys:

- `comp.set_background_color`
- `comp.set_renderer`
- `layer.create_text`
- `layer.set_transform`
- `effect.add_builtin`
- `effect.set_param`
- `property.set_expression`

Example mapping layer:

```go
type CapabilityMeta struct {
    Key      string
    Query    string
    Summary  string
}
```

The registry key is the stable recipe contract. `Query` is the current capindex
lookup string and may use Go API naming internally.

Example field-to-capability mappings:

- `comps[].background_color` -> `comp.set_background_color`
- `comps[].renderer` -> `comp.set_renderer`
- `comps[].layers[].type=text` -> `layer.create_text`
- `comps[].layers[].transform` -> `layer.set_transform`
- `comps[].layers[].effects[]` -> `effect.add_builtin`
- `comps[].layers[].effects[].params[]` -> `effect.set_param`
- `comps[].layers[].effects[].params[].expression.source` ->
  `property.set_expression`

At generation time, capability details should be joined from
`docs/capabilities.json` through the existing `internal/capindex` lookup. The
recipe registry owns stable recipe capability keys; the capability registry maps
those keys to capindex query strings. Capindex owns status, domain, verify
level, min version, boundary, and gate tests.

When a capability key has no mapping, or a mapped query is not found in
capindex, the drift test should fail unless the capability metadata explicitly
allows unknown.

## Documentation Rules

Recipe documentation should use recipe terms, not API terms:

- "field path" instead of "symbol"
- "recipe authoring" instead of "public API"
- "capability query" instead of "gate"
- "profile check" for `expected_profile`

Generated prose should be short. The primary value is accurate field coverage
and capability traceability, not long tutorials.

Fields should be grouped by object type and ordered in source struct order.
Nested object sections should be linked from parent fields.

Required fields should be described as two separate concepts:

- structural requiredness: no `omitempty` and non-pointer in the Go shape
- validation requiredness: required by `ValidateWithCapabilities`

Docs should avoid the plain label "required" unless it names which kind.
Defaultable fields must be documented as optional in the authoring sense even
when they are represented by non-pointer Go fields.

## Validation And Drift Gates

The implementation should add a test package for the generator with these
checks:

- Every included exported recipe field with a JSON tag appears in
  `recipe_schema.json`.
- Every field with a validation branch has a registry entry, starting with a
  curated first set rather than all branches.
- Every registry key references an existing generated recipe path.
- Every capability query in registry resolves against `docs/capabilities.json`,
  unless explicitly marked `allow_unknown`.
- `docs/recipe_schema.json` and `docs/recipe.md` are up to date.
- The generator output is deterministic.

Commands expected after implementation:

```powershell
go test ./cmd/recipedocgen -count=1
go test ./internal/recipe -count=1
go test ./...
go vet ./...
```

## Implementation Shape

Preferred package layout:

- `internal/recipedoc/model.go`
  - canonical field model and generated document model
- `internal/recipedoc/reflect.go`
  - reflection walker for recipe structs
- `internal/recipedoc/summary_registry.go`
  - field summaries by stable recipe path
- `internal/recipedoc/validation_registry.go`
  - validation notes, enum values, ranges, and required-if rules
- `internal/recipedoc/capability_registry.go`
  - field-path to stable recipe capability keys, and capability-key to capindex
    query mapping
- `internal/recipedoc/example_registry.go`
  - inline examples or example file references
- `internal/recipedoc/render_json.go`
  - deterministic JSON Schema-like rendering
- `internal/recipedoc/render_md.go`
  - markdown field reference rendering
- `cmd/recipedocgen/main.go`
  - CLI and file writing
- `cmd/recipedocgen/docs_uptodate_test.go`
  - drift gate

This mirrors the existing split between `cmd/docgen` and `internal/apidoc`,
but keeps recipe metadata separate from API annotation metadata.

## JSON Schema Shape

`docs/recipe_schema.json` should be JSON Schema-like rather than a custom
format invented from scratch.

The first version does not need to implement the full JSON Schema vocabulary,
but it should align with familiar field names where possible:

- `$schema`
- `$id`
- `title`
- `type`
- `properties`
- `items`
- `required`
- `enum`
- `description`
- custom extension fields under `x-aep-*`

Recipe-specific data should live under extension keys such as:

- `x-aep-path`
- `x-aep-structural-required`
- `x-aep-validation-required`
- `x-aep-capabilities`
- `x-aep-examples`

This keeps the output consumable by editors, LLM tooling, and future VSCode
integration without pretending to be a complete JSON Schema implementation.

## First-Slice Scope

The first implementation should document only the fields currently covered by
recipe validation and compile tests at a useful coarse granularity:

- top-level recipe and project fields
- composition basics and comp settings
- layer creation and common layer switches
- transform static values and transform keyframes
- effects and effect params
- shape primitives, stroke/fill, gradients, and shape filters
- text style fields
- `expected_profile` checks that already exist

If a supported struct field has no summary yet, the generator should still list
the field with type and path, but mark semantic metadata as missing. The drift
gate should allow this only for a temporary explicit allowlist that is visible
in the registry.

## Acceptance Criteria

The design is ready to implement when:

- The generated schema has stable recipe paths for every included field.
- The markdown reference can answer "what JSON can I write here?" without
  reading `internal/recipe`.
- Capability status shown in recipe docs is loaded from `docs/capabilities.json`,
  not copied by hand.
- Recipe capability keys are stable and are mapped to capindex queries in one
  registry.
- Running the generator twice produces identical output.
- Drift tests fail when a new recipe JSON field is added without doc coverage.
- Drift tests fail when a registry key references a removed or renamed recipe
  path.

## Next

Execute `plan.md` for the first implementation slice:

1. Build `internal/recipedoc` reflection model and tests.
2. Add split metadata registries for the highest-value fields.
3. Render JSON Schema-like `recipe_schema.json`.
4. Render `recipe.md`.
5. Add drift tests and wire `go generate`.

## Progress

Done:

- Spec written after recipe compiler/schema were split into smaller files.
- Boundary chosen: recipe docs are independent from API docs but can join
  capability details through capindex.
- Implementation plan written in `plan.md`.
- First implementation slice landed.
- Core capability metadata expanded for comp settings, layer switches/timing,
  text style fields, camera options, and light options; generated docs now join
  capability details for 99 recipe fields.

Current:

- Ready for shape/effect/expected-profile metadata expansion or editor
  integration.

## Open questions

- Whether `docs/recipe_index.json` is needed in the first implementation or can
  wait until an editor/agent integration consumes it.
- Whether example recipes should be copied into docs snippets or referenced by
  path only.
