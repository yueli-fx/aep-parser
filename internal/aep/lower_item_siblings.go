// internal/aep/lower_item_siblings.go
//
// Phase 2 Task 2.4 — formalize V2.1's "item sibling chunks" logic as a
// Phase 2 serializer primitive. Pure refactor: behavior unchanged from V2.1
// (still deep-clones from the AE 2020 dummy substrate); only the call shape
// moves to give V3 brainstorm a named primitive to reference.
//
// Background: AE expects each Item LIST in a Fold to be followed by 8
// sibling chunks (FEE LIST + fvdv / fiop / ftts / foac / fiac / fipc / fifl).
// Missing them → AE 25 reports "文件数据丢失" (per flightdeck/incident-reports/
// ae25-acceptance-gate.md). V2.1 deep-cloned them from the embedded dummy
// template inline in `Project.NewComposition`; Phase 2 extracts to give the
// primitive a name + dedicated file (Inv-2: one file, one responsibility).
//
// V2.2 escape hatch: substrate is bootstrap (Inv-6), not semantic template —
// chunks are deep-cloned per call so mutation doesn't leak.
package aep

import "github.com/example/aep-parser/internal/rifx"

// lowerItemSiblings returns deep-clones of the Fold-level sibling chunks AE
// expects after each Item LIST. ctx is reserved (V3 may key off capabilities);
// V2.2 ignores it — the AE 2020 substrate works on every supported AE
// version via back-compat.
func lowerItemSiblings(_ *lowerCtx) []*rifx.Chunk {
	ensureCompTemplate()
	out := make([]*rifx.Chunk, 0, len(compTmpl.siblingChunks))
	for _, sib := range compTmpl.siblingChunks {
		out = append(out, deepCloneChunk(sib))
	}
	return out
}
