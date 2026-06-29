// internal/aep/lower_item_siblings.go
//
// The "item sibling chunks" serializer primitive. Behavior unchanged from V2.1
// (still deep-clones from the AE 2020 dummy substrate); the call shape moves to
// give a named primitive to reference.
//
// Background: AE expects each Item LIST in a Fold to be followed by 8
// sibling chunks (FEE LIST + fvdv / fiop / ftts / foac / fiac / fipc / fifl).
// Missing them → AE 25 reports "文件数据丢失". V2.1 deep-cloned them from the
// embedded dummy template inline in `Project.NewComposition`; this extracts the
// primitive to its own file (one file, one responsibility).
//
// The substrate is bootstrap, not a semantic template — chunks are deep-cloned
// per call so mutation doesn't leak.
package serializer

import "github.com/yueli-fx/aep-parser/internal/rifx"

// lowerItemSiblings returns deep-clones of the Fold-level sibling chunks AE
// expects after each Item LIST. ctx is reserved (a future version may key off
// capabilities); the current implementation ignores it — the AE 2020
// substrate works on every supported AE version via back-compat.
func lowerItemSiblings(_ *lowerCtx) []*rifx.Chunk {
	ensureCompTemplate()
	out := make([]*rifx.Chunk, 0, len(compTmpl.siblingChunks))
	for _, sib := range compTmpl.siblingChunks {
		out = append(out, deepCloneChunk(sib))
	}
	return out
}
