// internal/aep/new_layer.go
//
// V2.2 Phase 3 — public-API entry for ShapeLayer creation. Atomic mutation
// pattern mirrors V2.1 Project.NewComposition (Inv-10 atomicity, Inv-11
// warnings-as-failure): lower runtime → chunk, snapshot mutable state,
// commit, rollback on any warning produced by downstream parse.
//
// Phase 3 wires only the construction side; parse-after-build full hydration
// (typed wrapper rebuilt from chunks) lands in Phase 4 roundtrip closure.
// For now the wrapper returned to the caller is the freshly constructed
// runtime instance, and comp.Layers carries the embedded *Layer base so V1
// readers (Layer.Name / Type / etc.) work.
package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// CompItemListForTest exposes a Composition's underlying Item LIST chunk
// for tests that need to inspect Item-level structure (Layr placement +
// Ewst siblings). Not part of public API.
func CompItemListForTest(c *Composition) *rifx.Chunk { return c.itemList }

// templateServiceLayerTypes are the LIST formTypes AE writes after the user
// Layr block — dummy template "service" layers carried through from the
// 2020_dummy_comp.aep template. AE expects user Layr to precede these.
// Each of these has its own ldta and consumes a LayerID — the user Layr
// must use an ID that doesn't collide with any of them (per iter 4 RE:
// LayerID collision with DLay → AE treats user Layr as deleted, comp shows
// layers.length=0). Template service layers occupy IDs 2..12 typically.
var templateServiceLayerTypes = map[rifx.ChunkID]bool{
	{'L', 'a', 'y', 'r'}: true, // user layers (counted to avoid collision among multiple NewShapeLayer)
	{'D', 'L', 'a', 'y'}: true,
	{'S', 'L', 'a', 'y'}: true,
	{'C', 'L', 'a', 'y'}: true,
	{'S', 'e', 'c', 'L'}: true,
}

// maxLayerIDInItemList scans every layer-list chunk in the comp's itemList
// (Layr / DLay / SLay / CLay / SecL) and returns the max LayerID found in
// their ldta @0x00. Returns 0 if no layer chunks exist.
func maxLayerIDInItemList(itemList *rifx.Chunk) uint32 {
	if itemList == nil {
		return 0
	}
	var maxID uint32
	for _, ch := range itemList.Children {
		if !ch.IsList() || !templateServiceLayerTypes[ch.FormType] {
			continue
		}
		ldta := ch.FindFirst(rifx.IDLdta)
		if ldta == nil || len(ldta.Data) < 4 {
			continue
		}
		id := uint32(ldta.Data[0])<<24 | uint32(ldta.Data[1])<<16 | uint32(ldta.Data[2])<<8 | uint32(ldta.Data[3])
		if id > maxID {
			maxID = id
		}
	}
	return maxID
}

// templateServiceLayerInsertTypes is templateServiceLayerTypes minus Layr —
// used by insertLayrPosition to find the first non-Layr template chunk.
var templateServiceLayerInsertTypes = map[rifx.ChunkID]bool{
	{'D', 'L', 'a', 'y'}: true,
	{'S', 'L', 'a', 'y'}: true,
	{'C', 'L', 'a', 'y'}: true,
	{'S', 'e', 'c', 'L'}: true,
}

// insertLayrPosition returns the index in children at which a new Layr
// chunk should be inserted. Per iter 4 bisection #2 finding: AE 2025
// rejects compositions where user Layr appears AFTER template service
// layers (DLay/SLay/CLay/SecL). Tolerance.aep order is:
//
//	[Item-header chunks] → Layr × N → DLay → SLay × 6 → CLay × 3 → SecL
//
// We insert at the position of the first service layer chunk (or end if
// none found, for builder configurations without a template).
func insertLayrPosition(children []*rifx.Chunk) int {
	for i, ch := range children {
		if ch.IsList() && templateServiceLayerInsertTypes[ch.FormType] {
			return i
		}
	}
	return len(children)
}

// NewShapeLayer adds a new empty ShapeLayer to the composition.
//
// Required:
//
//	name — non-empty string (matches NewComposition validation contract)
//
// Returns the typed *ShapeLayer wrapper; the embedded *Layer is also
// appended to comp.Layers so V1 lookup paths (Composition.LayerByID /
// LayerByName) work immediately. ID is auto-assigned via the project's
// monotonic item-ID counter (Inv-9 — never reused; layer IDs share the
// item-ID namespace per V1 parser convention).
//
// Atomic mutation (Inv-10): if lowering fails, or downstream parse emits
// any warning, all state mutated by this call is rolled back to the
// pre-call snapshot before the error is returned.
func (c *Composition) NewShapeLayer(name string) (*ShapeLayer, error) {
	if name == "" {
		return nil, fmt.Errorf("ShapeLayer name cannot be empty")
	}
	if c.proj == nil {
		return nil, fmt.Errorf("internal: comp has no project back-ref")
	}
	if c.itemList == nil {
		return nil, fmt.Errorf("internal: comp has no itemList chunk")
	}

	// 1. Construct runtime ShapeLayer (Phase 1 types). Layer IDs occupy a
	//    per-composition namespace (NOT the project-wide Item ID namespace —
	//    iter 4 RE: template service layers DLay/SLay/CLay/SecL hold IDs 2..12
	//    in the comp's itemList; user Layr ID must not collide, otherwise
	//    AE treats the user Layr as deleted via the DLay ID match).
	layerID := maxLayerIDInItemList(c.itemList) + 1
	base := &Layer{
		Type: LayerTypeShape,
		Name: name,
		ID:   layerID,
		comp: c,
		back: &layerBackrefs{},
	}
	// Bump project nextItemID so head-chunk counter sync (write.go::
	// syncHeadCounters) covers our layer ID. AE 2025 validates head counter
	// >= max(item/layer IDs) and silently drops layers above it (iter 4 RE).
	if layerID >= c.proj.nextItemID {
		c.proj.nextItemID = layerID + 1
	}

	// Bump cdta @0x18 (cdtaSecondaryDivisor18) from 600 (fresh-comp marker)
	// to TickRate. cdta_layout.go: "AE rewrites to TickRate on user mod".
	// iter 4 RE: tolerance.aep (which AE 2025 accepts as a real comp with
	// real user layers) has this field = TickRate, our fresh comp has 600.
	// Hypothesis: AE uses this as a "comp has user content" gate.
	if c.cdta != nil && len(c.cdta.Data) >= cdtaSecondaryDivisor18+4 {
		tr := uint32(c.TickRate)
		if tr > 0 {
			c.cdta.Data[cdtaSecondaryDivisor18+0] = byte(tr >> 24)
			c.cdta.Data[cdtaSecondaryDivisor18+1] = byte(tr >> 16)
			c.cdta.Data[cdtaSecondaryDivisor18+2] = byte(tr >> 8)
			c.cdta.Data[cdtaSecondaryDivisor18+3] = byte(tr)
		}
	}
	s := WrapShapeLayer(base)

	// 2. Lower runtime → LIST(Layr) chunk (Phase 2).
	ctx := &lowerCtx{
		tickRate:     c.TickRate,
		compDuration: c.Duration,
		capabilities: Capabilities(c.proj.target),
		nextLayerID:  c.proj.allocItemID,
	}
	layrChunk, err := lowerShapeLayer(s, ctx)
	if err != nil {
		return nil, fmt.Errorf("lower ShapeLayer: %w", err)
	}

	// 3. Snapshot pre-mutation state (Inv-10 atomicity). For itemList we
	//    save a slice copy (since the insert is mid-slice now, not append-end).
	oldItemChildren := append([]*rifx.Chunk(nil), c.itemList.Children...)
	oldLayersLen := len(c.Layers)
	oldWarningsLen := len(c.proj.Warnings)

	// 4. Commit: insert Layr (+ its Ewst sibling) into itemList BEFORE
	//    first template service layer (DLay/SLay/CLay/SecL) — AE 2025
	//    rejects when user Layr appears after these (iter 4 bisection #2).
	//    Multiple user Layr additions stack in insertion order before
	//    the service block.
	//
	//    iter-5 RE finding: every AE-saved Layr in the Item LIST is
	//    immediately followed by an empty LIST(Ewst, 0 children) sibling.
	//    Template service layers carry their own Ewst (already in the
	//    template). User-built Layrs must emit one too — without it, AE
	//    2025 silently drops the layer at instantiation stage (variant #2
	//    empty ShapeLayer reproduces this even with zero shape kids).
	base.back.layrList = layrChunk
	base.shapeDirty = true // gate for syncShapeLayerChunks
	ewstSibling := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDEwst}
	insertIdx := insertLayrPosition(c.itemList.Children)
	// Grow slice by 2 (Layr + Ewst) and shift any existing tail two slots.
	c.itemList.Children = append(c.itemList.Children, nil, nil)
	copy(c.itemList.Children[insertIdx+2:], c.itemList.Children[insertIdx:len(c.itemList.Children)-2])
	c.itemList.Children[insertIdx] = layrChunk
	c.itemList.Children[insertIdx+1] = ewstSibling
	c.Layers = append(c.Layers, base)

	// 5. Warnings-as-failure (Inv-11): if any warnings appeared, rollback.
	// Phase 3 doesn't re-parse, so this is a defensive guard for Phase 4 to
	// rely on (re-parse closed loop arrives there).
	if len(c.proj.Warnings) != oldWarningsLen {
		c.itemList.Children = oldItemChildren
		c.Layers = c.Layers[:oldLayersLen]
		newWarnings := append([]string(nil), c.proj.Warnings[oldWarningsLen:]...)
		c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: NewShapeLayer produced %d parser warning(s): %v", len(newWarnings), newWarnings)
	}

	return s, nil
}
