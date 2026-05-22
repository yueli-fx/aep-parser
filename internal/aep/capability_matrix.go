// internal/aep/capability_matrix.go
package aep

// AECapabilities describes the serializer-affecting traits of a target AE
// version. V2.2 ships with this struct deliberately EMPTY — the escape hatch
// (AE 2020 canonical minimum, see spec §1.4 / §4.6) covers all currently
// known cross-version differences without requiring lowering to branch.
//
// Adding a trait requires the §1.4 admission rule to hold:
//
//  1. cross-version observable difference (measured, not theorized);
//  2. escape hatch single canonical cannot cover both versions; AND
//  3. lowering must actually branch on the trait.
//
// All three conditions must be met. Phase 0 RE-S9 evaluated seven candidates
// and none qualified for V2.2 — listed here for §6.5 traceability:
//
//   - LdtaSize             (160 AE 2020/22 vs 164 AE 25 zero-pad tail)
//   - FEEHasPpSn           (FEE LIST ppSn child absent in AE 2020)
//   - MaterialLightingGroups (10 extra Material groups in AE 2025)
//   - TdgpDefaultChildren  (19 in AE 2020/22 vs 37 in AE 2025)
//   - ShapeMatchNameVariants (RE-S9: cross-version byte-identical)
//   - PathBezierEncoding   (RE-S9: shap/shph/lhd3/ldat cross-version identical)
//   - KeyframeEaseEncoding (RE-S9: cosmetic flag-byte diff only; AE 2025 reads AE 2020 ldat fine)
//
// V3 brainstorm may revisit (OQ-1 capability auto-derive). Keeping the struct
// + lookup function reserves the API surface so a future trait addition is
// non-breaking to call sites that already do `caps := Capabilities(target)`.
type AECapabilities struct {
	// V2.2: empty by design. Add fields only when §1.4 admission rule holds.
}

// Capabilities returns the capability set for the given AE target. Pure
// function lookup — intentionally NOT a method on *Project so a future V3
// auto-derive path (OQ-1) can plug in without disturbing call sites.
//
// V2.2: returns the same empty struct for every target.
func Capabilities(target AETarget) AECapabilities {
	_ = target
	return AECapabilities{}
}
