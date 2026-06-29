# Index — scene-architecture

## State

This work package tracks architecture cleanup for `internal/scene` and the
serializer bridge. It was opened after reviewing `scene_project.go`,
`scene_layer.go`, property groups, shape graph, text writers, writer
interfaces, warning flow, and mutation rollback patterns.

Breaking changes are allowed during this stage, but they should be justified by
long-term structural value. Do not preserve public API shape at the cost of
ongoing state inconsistency, but also do not break callers for cosmetic moves.

## Findings

- Project-level persistent ID indexes are unsafe as a first step while
  `Project.Compositions`, `Project.Footage`, and `Composition.Layers` remain
  public mutable slices.
- Local per-call indexes are safe for source-heavy composition views because
  they cannot become stale after the call returns.
- `Layer` is too large, but its exported fields are the public scene model.
  Private runtime/writeback state should be collected first through
  `scene_wiring.go`.
- `Warnings []string` is part of mutation transaction semantics. Many
  structural mutators snapshot `len(Project.Warnings)` and rollback on new
  warnings, so structured warnings must be additive and transaction-aware.
- `scene_shape_graph.go` and `serializer/lower_shape_node.go` are the largest
  maintainability risks. They should be mechanically split by shape-node family
  before further behavior expansion.

## Execution Order

1. Optimize source-heavy composition view methods with local indexes.
2. Group `Layer` private runtime/writeback state without changing public fields.
3. Mechanically split shape graph and shape lowering files by node family.
4. Design additive structured warnings while preserving `Warnings []string`
   rollback behavior.
5. Revisit public model breaking changes once the private-state and shape-file
   splits reduce the blast radius.

## Progress

- [x] Audited scene/serializer architecture hotspots.
- [x] Confirmed persistent Project indexes are not safe with current public
      mutable slices.
- [x] Added local source indexes for composition source views.
- [ ] Group Layer private state.
- [ ] Split shape graph / shape lowering files.
- [ ] Design structured warning migration.

## Verification

Current first slice:

- `go test ./internal/scene -count=1`
