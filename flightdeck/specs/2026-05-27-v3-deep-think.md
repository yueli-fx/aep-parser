# V3 deep think — open questions, risk register, migration strategy

**Status**: brainstorm. Not a plan. Builds on `2026-05-22-v3-direction.md` (M1-M8 framework) with the questions that need answers before writing the plan.

**Created**: 2026-05-27 after P2 parity 收尾 (PASS 259).

**Reader**: pick `## TL;DR` if you want the bottom line; the rest unpacks the trade-offs.

---

## TL;DR — what V3 has to decide

1. **Migration shape**: in-place refactor (incremental, low-risk, less-clean separation) vs greenfield `internal/scene/` + adapter layer (clean cut, more code, slower). **Recommend in-place** — single package is in CLAUDE.md hard constraints; greenfield breaks it without payback proportional to disruption.
2. **Eager vs lazy chunk writes**: today every `SetX` mutates RIFX bytes immediately. V3 ideal is "scene is source of truth, chunks regenerated on `WriteAEP()`". **Recommend keeping eager for known fields, lazy for V3-only structural ops.** Don't migrate the eager path away — it's our length-preserving guarantee.
3. **Opaque preservation**: chunks we don't understand must survive WriteAEP byte-identical. V3 scene types need an `opaque` shard per layer/comp/property. **This is non-negotiable for the AE 2025 ship gate**.
4. **Capability matrix scope**: static map per target is fine for serialization decisions. Don't try to encode every per-version chunk-layout quirk; promote to capability only when we have ≥ 2 call sites diverging on it.
5. **Starting slice**: `LayerGraph` (Layer / Effect / Mask / Property as scene types, with serializer producing today's chunk layout). Defer ShapeGraph (M3) and EffectSchema (M4 generic container) — those are big sub-projects.

---

## 1. Validating M1-M8 against current state

### M1 Runtime Object Model — needs adjustment

Direction says runtime types hold *no chunk refs*. Today's types reference `*rifx.Chunk` extensively (e.g. `Property.cdat / Property.ldat / Property.tdb4 / Layer.ldta`). These power length-preserving writes.

**Adjustment**: V3 runtime types retain a `back *backrefs` shard pointing to the underlying chunks for length-preserving writes. The back-ref is **purely an optimization** — purging it means "regenerate on next WriteAEP". The logical surface (`Position() *PropertyStream`, etc.) doesn't expose chunk refs.

```go
type Property struct {
    // Logical (V3 surface):
    Dimensions    int
    StaticValue   []float64
    Keyframes     []Keyframe
    Expression    string

    // Serialization shard (private; nil for built-from-scratch properties):
    back *propertyBackrefs
}

type propertyBackrefs struct {
    tdbs, tdb4, cdat, ldat, lhd3, tdsb, tdum, tduM *rifx.Chunk
    opaque map[rifx.ChunkID]*rifx.Chunk  // chunks we recognize but don't decode
}
```

V1 SetX continues to work via `back` — fast path, length-preserving. New structural ops (Remove / Duplicate) go through serializer regeneration.

### M2 Archetype minimization — agree, but with one caveat

V3 doc wants minimal archetypes (one per layer kind). Caveat: AE rejects "minimal" too easily — we hit this with V2.1 / V2.2 ship gates. The actual archetype is whatever the AE 2020 / 2025 acceptance gates demand.

**Build process**:
1. RE one AE-saved fixture per layer kind (Shape / Text / AV / Camera / Light / Solid)
2. Diff against the byte-minimum that AE still loads
3. The minimum is the archetype

This is RE-heavy and assumes user can produce fixtures. Realistic timeline: 1 fixture per session.

### M3 Shape graph — biggest sub-project, deferred from Phase 1

V2.2 ShapeLayer alpha has known silent-drop limitations (Ellipse / Path / Stroke / per-group transforms / runtime-only Layer Transform). V3's Shape graph needs to **solve** these. That means:

- Embed boilerplate bytes for every primitive we want to write (today: Rect + Fill only)
- RE Ellipse / Path / Stroke / GradientFill / GradientStroke / Trim / Merge / Repeater / each shape effect
- Each is its own RE cycle

**Recommendation**: ShapeLayer V2.2 stays alpha; V3 Phase 1 doesn't unify. ShapeGraph becomes V3 Phase 5 (after layer graph + property stream IR + capability matrix are stable).

### M4 Effect generic container — agree, but opaque round-trip is the hard part

Generic effect = `{MatchName, Params}`. The hard part is opaque round-trip when the parser doesn't know the param schema. Today, parameters are decoded from `tdgp/tdbs/cdat` byte-by-byte; that works for any effect because the chunk tree is generic.

V3 keeps the generic decoder. The "schema lookup" (e.g. "Blurriness is param 0001") is **already in py-aep's `specs.py`** (2353 lines). Either:
- (a) port py-aep's table verbatim (1-time investment, multi-thousand-line table)
- (b) leave names as `ADBE Gaussian Blur 2-0001` — user looks up the index, our API just exposes index-keyed access

**Recommendation**: (b). Schema port is busywork — better keep that out of V3 core and offer as a separate sidecar package later.

### M5 PropertyStream IR — mostly converges with current model

Today's `Property` already has Dimensions / Keyframes / StaticValue / Expression / ExpressionEnabled. The biggest gap is **time unit**: today `Keyframe.Time` is in seconds with `tickRate` field used by setters. V3 wants seconds-only at runtime, TickRate at serializer.

**Migration path**: keep `Keyframe.Time` in seconds (we already do). Drop the runtime tickRate hint when the back-ref is removed — serializer will look up the comp's TickRate at write time.

### M6 Graph mutation API — needs careful design

Listed ops:
- `project.CreateLayer / DeleteLayer / DuplicateLayer / MoveAfter / Reparent`
- `layer.InsertEffect / RemoveEffect`
- `shapeLayer.AttachNode`
- `project.CloneSubgraph`

Each is structural. Each needs:
1. Logical-side mutation (update slices, validate refs)
2. Serializer-side regeneration (rewrite the affected chunk subtree)
3. Atomic semantics (failure → roll back)

Today `NewComposition` and friends use the "atomic mutation + warnings rollback" pattern (CLAUDE.md Invariant #10/#11). V3 mutations should reuse that pattern wholesale.

### M7 Capability matrix — start small

V3 doc lists 7 candidate capabilities. Today V2.1 already touches:
- `LdtaSize` (160 vs 164)
- `FEEHasPpSn` (AE 2022+)
- `TdgpVariant` (19 vs 37 children)

That's the practical capability set today. Add more only when a serializer site actually needs to branch on AE version. Premature capabilities = config sprawl.

### M8 Serializer boundary — split, but stay in `internal/aep/`

CLAUDE.md hard constraint #3: single `internal/aep` package. Sub-packages would force a public re-export layer that ripples through every test.

**Compromise**: keep `internal/aep/` as the single package, but reorganize files:
- `scene_*.go` — V3 logical types (no chunk refs in public fields)
- `serialize_*.go` — chunk synthesis (writes scene → chunks)
- `parse_*.go` — chunk parsing (reads chunks → scene; already partitioned)
- `back_*.go` — back-ref shards + length-preserving fast paths

No new packages. Boundary enforced by file naming + lint rules + scar (write a scar documenting "do not call serialize_* from scene_*").

---

## 2. Open questions (need answer before plan)

### Q1: Hard break or soft migration of public API?

Today's public surface (used by external consumers, also by every test):
- `aep.Open(path) (*Project, error)`
- `proj.Compositions []*Composition`
- `layer.Properties []*Property`
- `layer.Position() *Property`
- `prop.SetStaticValue(v) error`
- ...

V3 ideal might rename to `aep.Open() (*Scene, error)` etc. But:
- 259 tests pass on the current API
- CLAUDE.md hard constraint #2: "public API not动"

**Recommended answer**: soft migration. Same names, same return types. The scene types **become** the current types — just refactored internally. `aep.Open` returns `*aep.Project` which is now the scene Project. New methods get added (`proj.CreateLayer` etc.); old methods preserved.

### Q2: How to handle "open then modify then write" vs "scratch then write"?

Today:
- `aep.Open` → tree of chunks + scene wrappers; `SetX` is fast (in-place byte write); `WriteAEP` serializes the tree (mostly verbatim) + recomputed sizes for length-variable splices.
- `aep.NewProject` → empty tree from a template; mutations append; `WriteAEP` serializes the tree.

Both paths converge at "tree of chunks". V3 doesn't have to change this if we accept: **the chunk tree is always present as serialization back-ref**. Even "from scratch" goes through an empty-template chunk tree.

**Recommended answer**: keep the chunk tree as the always-present serialization shard. Scene types layer on top. No mode switching ("are we in scratch mode or open mode?") — there's only one mode.

### Q3: Where does each property live — is it owned by Layer, or by a separate property tree?

Today: `Layer.Properties []*Property` (flat), plus the V2.2-introduced `AEPropertyGroup` tree, plus `Layer.Effects[]` with their own `[]*Property`.

The flat list is convenient for `layer.Position()` style lookup. The tree is what AE's data model has. The effect-side lists are what users actually iterate.

**Three-way redundancy** that should converge in V3. But the convergence is invasive — every existing test references one of these surfaces.

**Recommended answer**: keep all three surfaces in V3 (P2c shipped the tree explicitly to add the hierarchy without breaking the flat list). Mark the tree as canonical; flat list is a generated view. Pointer-identity guarantees mutations propagate.

### Q4: Effect parameters — by index, by match-name, by symbolic name?

Today: `effect.Parameters []*Property`. Match-name like `ADBE Gaussian Blur 2-0001` (the "0001" is parameter index). Users iterate.

py-aep: `effect.parameter("Blurriness")` (symbolic). Requires schema port.

**Recommended answer**: keep index-based. Add `effect.Parameter(matchName)` helper that does a linear search. Symbolic names are a v.next thing.

### Q5: What about chunks we never decode (e.g. `EwSt` boilerplate, `Pefl` payload internals, EG controllers)?

Today: parser stores them as `rifx.Chunk` somewhere reachable; WriteAEP emits them verbatim.

V3 (if we ever fully detach chunk refs): they'd be lost on regeneration.

**Recommended answer**: scene types carry an `opaque []*rifx.Chunk` shard for unhandled siblings/children. Serializer re-emits them in original position. This is the only way the AE acceptance gate stays green for the chunks we haven't reverse-engineered.

### Q6: Threading / concurrency — does V3 change anything?

Today (CLAUDE.md hard constraint): "all Set* mutate shared chunk bytes; caller locks". Scar `concurrency-unsafe-shared-chunk-bytes.md` documents this.

V3 logical mutations are pure value updates on scene types. **Better story**: scene-level locking is easier than chunk-byte-level locking. We can promise "Project + everything reachable through scene API is safe for one-writer / many-reader as long as readers don't observe partially-written state". Caller still locks across multi-step mutations.

**Recommended answer**: keep current model (caller locks). Don't promise more.

### Q7: How does write-back work when both eager and lazy paths coexist?

Scenario: user opens project, calls `layer.SetVisible(true)` (eager byte flip), then calls `comp.AddLayer(newLayer)` (lazy, needs regeneration). On `WriteAEP`, we need to:
- Preserve the eager byte change
- Regenerate only the comp-children-that-changed (atomic insertion)

**Recommended answer**: serialize from chunk tree (current behavior). Eager writes already updated the chunks. Lazy ops also update the chunks (build a new subtree, splice in). `WriteAEP` is just `root.Write(w)` — no "mode" required. The serializer's job is to keep the chunk tree consistent with the scene model when scene-level structural mutations happen.

---

## 3. Risk register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| AE 2025 silent-drop of V3-synthesized layers (like V2.2 ShapeLayer) | High for new layer kinds | High — blocks ship gate | Per-layer-kind ship gate before merging that slice |
| Public API drift breaks downstream consumers | Medium — we have no known consumers but CLAUDE.md treats it as inviolable | Medium | Keep same names; new methods only; soft migration |
| Opaque preservation misses a chunk → AE silent-drop | High during dev | High | Round-trip diff test on every fixture before each commit |
| Capability matrix sprawl | Medium | Low — code smell, not breakage | Lint rule: capability fields must have ≥ 2 serializer call sites |
| Mutation API atomic-rollback edge cases (warnings emitted during regen) | Medium | High — leaves project in inconsistent state | Reuse `Project.Warnings` invariant pattern (V2.1 proven) |
| Effect schema port pressure as users hit unknown effects | Low if we say "no" up-front | Low | Document early: index-keyed access, schema port is out of V3 scope |
| Migration timeline slips → unmaintained V2 alpha + half-built V3 | High | High | Phase boundaries are commit-able shippable products, not "WIP" |

---

## 4. Recommended starting slice — V3 Phase 1

**Goal**: detach `Property.tdb4 / Property.cdat` etc. from `Property` struct top-level, move into `back *propertyBackrefs`. Public API surface unchanged. Tests still pass byte-identical.

**Concretely**:
1. Add `propertyBackrefs` struct holding what's today directly on `Property`
2. Migrate accessor methods to read through `p.back.tdb4` etc.
3. Same for `Layer.ldta` → `Layer.back.ldta` etc.
4. Run full test suite — every test must pass unchanged

This is **pure refactoring**. No new behavior. The payoff is:
- Clear scene/serialization boundary at the struct level
- Future V3 phases can mutate scene fields freely; back-ref is opt-in for length-preserving writes
- A clean place to add the `opaque []*rifx.Chunk` field

**Phase 1 acceptance criteria**: 259 PASS, byte-identical WriteAEP on all RE fixtures, no public-API delta.

**Estimated work**: 1-2 sessions of refactoring. Low risk, foundational.

**Phase 2 candidate** (after Phase 1): `Composition.DeleteLayer(idx)` — first structural mutation through the scene + back-ref split. Needs RE on AE acceptance (does AE accept comp with layer removed via splice?). Bisect against AE 2025 ship gate.

**Phase 3+**: layer graph ops (Insert / Move / Reparent), capability matrix expansion, shape graph V3 (subsumes V2.2 alpha).

---

## 5. What V3 explicitly will NOT do

(carried forward from `2026-05-22-v3-direction.md`, re-confirmed):

- Effect parameter schema library (1000+ effects). Index-keyed access stays.
- Expression engine / evaluation. Strings stay strings.
- Render queue / output module. AE-runtime concepts.
- Color management deep integration. CMS chunk stays as JSON Utf8.
- ExtendScript 1-based indexing. We stay 0-based Go.
- py-aep-style pythonic descriptors. Idiomatic Go pair-getters/setters stay.

---

## 6. Decision asks (for user / next session)

Before writing the V3 Phase 1 plan, want explicit answer on:

1. **Q1 (soft migration)** — confirm "same public types, internal refactor only"?
2. **Q5 (opaque preservation shard)** — confirm we add `opaque []*rifx.Chunk` to scene types now, not later?
3. **Phase 1 scope** — confirm "back-ref struct extraction only, no behavior change"? Or include something user-visible to make the phase feel like progress?
4. **Timeline expectations** — comfortable with multi-week V3 timeline (per `2026-05-22-v3-direction.md` projection)?

If all four answers are yes/yes/yes/yes, the next step is `writing-plans` → `flightdeck/flight-plans/2026-05-28-v3-phase1-backrefs-plan.md`.

---

## Related

- `2026-05-22-v3-direction.md` — M1-M8 framework (this builds on it, doesn't replace it)
- `2026-05-26-py-aep-parity-design.md` — what V3 inherits (P2 fully ✅, remaining ❌ all V3/P3)
- `CLAUDE.md` § 硬约束 — length-preserving + public API + single package constraints V3 must respect
- `scars/ae25-acceptance-gate.md` — the empirical ground for "AE silently drops things we don't synthesize correctly"
