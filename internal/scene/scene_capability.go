// internal/aep/capability_matrix.go
package scene

import "github.com/example/aep-parser/internal/codec"

// AECapabilities describes the serializer-affecting traits of a target AE
// version. V2.2 ships with this struct deliberately EMPTY — the escape hatch
// (AE 2020 canonical minimum) covers all currently
// known cross-version differences without requiring lowering to branch.
//
// Adding a trait requires the admission rule to hold:
//
//  1. cross-version observable difference (measured, not theorized);
//  2. escape hatch single canonical cannot cover both versions; AND
//  3. lowering must actually branch on the trait.
//
// All three conditions must be met. Seven candidates were evaluated.
//
// ADMITTED LdtaSize: the AE 2020 ShapeLayer ship-gate disproved the
// escape-hatch single-canonical assumption. AE 2020's ldta reader treats a
// 164-B ShapeLayer ldta as corrupt and silently skips the layer ("项目文件似乎
// 已损坏（跳过部分：1）"); AE 2025 reads its native 164 B. All three
// conditions now hold — measured difference (AE-2020-native ldta = 160 B,
// AE-2025-native = 164 B), single canonical cannot cover both, lowering
// branches in buildLdtaBytes.
//
// The remaining six candidates still don't qualify for V2.2 — listed
// here for traceability:
//
//   - FEEHasPpSn           (FEE LIST ppSn child absent in AE 2020)
//   - MaterialLightingGroups (10 extra Material groups in AE 2025)
//   - TdgpDefaultChildren  (19 in AE 2020/22 vs 37 in AE 2025)
//   - ShapeMatchNameVariants (cross-version byte-identical)
//   - PathBezierEncoding   (shap/shph/lhd3/ldat cross-version identical)
//   - KeyframeEaseEncoding (cosmetic flag-byte diff only; AE 2025 reads AE 2020 ldat fine)
//
// V3 brainstorm may revisit (capability auto-derive). Keeping the struct
// + lookup function reserves the API surface so a future trait addition is
// non-breaking to call sites that already do `caps := Capabilities(target)`.
type AECapabilities struct {
	// LdtaSize is the ShapeLayer ldta payload size (bytes) the target AE
	// version accepts: 160 for AE 2020/2022, 164 for AE 2025. Emitting the
	// wrong size makes AE 2020 reject the layer as corrupt. 0 is never
	// returned by Capabilities (buildLdtaBytes falls back to 164 if unset).
	LdtaSize int
}

// Capabilities returns the capability set for the given AE target. Pure
// function lookup — intentionally NOT a method on *Project so a future V3
// auto-derive path can plug in without disturbing call sites.
func Capabilities(target AETarget) AECapabilities {
	c := AECapabilities{LdtaSize: codec.LdtaSize2020}
	if target >= TargetAE2025 {
		c.LdtaSize = codec.LdtaSize2025
	}
	return c
}
