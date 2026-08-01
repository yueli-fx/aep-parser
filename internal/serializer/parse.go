package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// Open parses an .aep file by path and returns the Project.
//
// The file is read into bounded memory first and parsed from a bytes.Reader:
// rifx.readChunk does many small Read+Seek calls per chunk, so handing it the
// raw *os.File issues thousands of syscalls (≈97% of parse time on an 8 MB
// project). An in-memory reader turns those into pointer moves — ~100× faster.
func Open(path string) (*Project, error) {
	return OpenWithLimits(path, rifx.DefaultLimits)
}

// OpenWithLimits parses an .aep path with explicit RIFX resource budgets.
func OpenWithLimits(path string, limits rifx.Limits) (*Project, error) {
	maxInputBytes := limits.MaxInputBytes
	if maxInputBytes == 0 {
		maxInputBytes = rifx.DefaultLimits.MaxInputBytes
	}
	data, err := readFileBounded(path, maxInputBytes)
	if err != nil {
		return nil, fmt.Errorf("aep: open %q: %w", path, err)
	}
	return FromReaderWithLimits(bytes.NewReader(data), limits)
}

func readFileBounded(path string, maxBytes uint64) ([]byte, error) {
	if maxBytes > uint64(math.MaxInt64-1) {
		return nil, fmt.Errorf("max input bytes %d exceeds supported file-read limit %d", maxBytes, int64(math.MaxInt64-1))
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)+1))
	if err != nil {
		return nil, err
	}
	if uint64(len(data)) > maxBytes {
		return nil, &rifx.LimitError{Resource: "input bytes", Limit: maxBytes, Actual: uint64(len(data))}
	}
	return data, nil
}

// FromReader parses an .aep file from an io.ReadSeeker.
func FromReader(r io.ReadSeeker) (*Project, error) {
	return FromReaderWithLimits(r, rifx.DefaultLimits)
}

// FromReaderWithLimits parses an .aep stream with explicit RIFX resource budgets.
func FromReaderWithLimits(r io.ReadSeeker, limits rifx.Limits) (*Project, error) {
	root, err := rifx.ParseWithLimits(r, limits)
	if err != nil {
		return nil, fmt.Errorf("aep: parse RIFX: %w", err)
	}
	proj, err := parseProject(root)
	if err != nil {
		return nil, err
	}
	if pb := projectBack(proj); pb != nil {
		pb.root = root
	}
	return proj, nil
}

// Reopen serializes the project to memory and re-parses the bytes, returning
// the fresh *Project. Round-tripping upgrades layers built by the structural
// New* APIs into fully parsed layers (property tree + chunk back-refs), which
// unlocks the parsed-layer-only write paths (AddEffect parade auto-create,
// Camera*/Light* setters, ...) on them.
// (Full contract lives on the aep.Reopen facade — docgen source.)
func Reopen(p *Project) (*Project, error) {
	if p == nil {
		return nil, fmt.Errorf("aep: Reopen: project is nil")
	}
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		return nil, fmt.Errorf("aep: Reopen: %w", err)
	}
	return FromReader(bytes.NewReader(buf.Bytes()))
}

func parseProject(root *rifx.Chunk) (*Project, error) {
	pb := &projectBackrefs{}
	proj := &Project{}
	scene.SetProjectBack(proj, pb)

	// Project-level header chunks: nhed (32-byte) + nnhd (40-byte) sit
	// as direct children of the root LIST/RIFX. Both carry a copy of
	// BitsPerChannel at the same byte position relative to chunk start.
	if nhed := root.FindFirst(rifx.IDNhed); nhed != nil {
		pb.nhedChunk = nhed
		if len(nhed.Data) > 0x0F {
			proj.BitsPerChannel = BitsPerChannel(nhed.Data[0x0F])
		}
	}
	if nnhd := root.FindFirst(rifx.IDNnhd); nnhd != nil {
		pb.nnhdChunk = nnhd
	}

	// Project-level setting chunks. All sit as direct root children — capture
	// refs for the Set* methods in project_settings.go.
	for _, c := range root.Children {
		switch c.ID {
		case rifx.IDAcer:
			pb.acerChunk = c
		case rifx.IDAdfr:
			pb.adfrChunk = c
		case rifx.IDDwga:
			pb.dwgaChunk = c
		}
		if c.IsList() {
			switch c.FormType {
			case rifx.IDGpuG:
				if k := c.FindFirst(rifx.IDUtf8); k != nil {
					pb.gpugUtf8 = k
				}
			case rifx.IDExEn:
				if k := c.FindFirst(rifx.IDUtf8); k != nil {
					pb.exenUtf8 = k
				}
			}
		}
	}

	// CMS settings JSON (AE 24+) — stored as a Utf8 chunk containing JSON.
	// Identified by the presence of "lutInterpolationMethod" in the content.
	for _, c := range root.Children {
		if !c.IsList() && c.ID == rifx.IDUtf8 {
			content := string(c.Data)
			if len(content) > 0 && (content[0] == '{' || content[0] == '[') {
				// Looks like JSON, check for CMS markers
				if strings.Contains(content, "lutInterpolationMethod") || strings.Contains(content, "colorManagementSystem") {
					pb.cmsUtf8 = c
					break
				}
			}
		}
	}

	// Project items live as direct Item children of the root Fold and nested
	// folder Sfdr containers. Traverse only those containers so the parent
	// relationship and mixed sibling order remain intact.
	if err := parseProjectItems(root, proj); err != nil {
		return nil, err
	}
	initDerived(proj, root)
	parseRenderQueue(root, proj)
	return proj, nil
}

// parseProjectItems builds both the type-specific payload indexes and the
// canonical project-panel topology from the same traversal.
func parseProjectItems(root *rifx.Chunk, proj *Project) error {
	rootFold := root.FindFirstList(rifx.IDFold)
	if rootFold == nil {
		return nil
	}
	proj.Items = proj.Items[:0]
	return walkProjectItems(rootFold, 0, func(item *rifx.Chunk, entry ProjectItem) error {
		if err := parseItem(item, proj); err != nil {
			return err
		}
		proj.Items = append(proj.Items, entry)
		return nil
	})
}

// rebuildProjectItems refreshes the topology after a structural mutation.
// The raw Fold/Sfdr tree is authoritative, so mutation code does not maintain
// a second incremental copy of parent/order state.
func rebuildProjectItems(proj *Project) error {
	pb := projectBack(proj)
	if pb == nil || pb.rootFold == nil {
		return fmt.Errorf("project has no root Fold back-ref")
	}
	items := make([]ProjectItem, 0, len(proj.Items))
	if err := walkProjectItems(pb.rootFold, 0, func(_ *rifx.Chunk, entry ProjectItem) error {
		items = append(items, entry)
		return nil
	}); err != nil {
		return err
	}
	proj.Items = items
	return nil
}

// walkProjectItems visits project items in panel preorder. Order is local to
// each Fold/Sfdr container and counts every recognized project item sibling.
func walkProjectItems(container *rifx.Chunk, parentID uint32, visit func(*rifx.Chunk, ProjectItem) error) error {
	order := 0
	for _, child := range container.Children {
		if !isItemList(child) {
			continue
		}
		kind, id, err := classifyItem(child)
		if err != nil {
			return err
		}
		itemOrder := order
		order++
		if kind == ItemTypeUnknown {
			continue
		}
		entry := ProjectItem{ID: id, Kind: kind, ParentID: parentID, Order: itemOrder}
		if err := visit(child, entry); err != nil {
			return err
		}
		if kind == ItemTypeFolder {
			if sfdr := child.FindFirstList(rifx.IDSfdr); sfdr != nil {
				if err := walkProjectItems(sfdr, id, visit); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// initDerived 在 parseProject 收尾时调用，初始化 Project 的 derived state：
//   - nextItemID = max(已有所有 item IDs) + 1（monotonic counter for NewComposition / Duplicate）
//     Must include LAYER IDs too — AE often assigns layer.id > footage.id within
//     a comp, so excluding layers can leave nextItemID below an in-use layer ID
//     (collision on next allocItemID, surfaced on AE 23+ matte fixtures where
//     layer IDs run higher than any folder/comp/footage).
//   - rootFold = root Egg! 下第一个 formType=Fold 的 LIST（cached for V2 mutations）
//
// 参数 rifxRoot 是 parseProject 顶层 *rifx.Chunk（formType=Egg!）。
// Free function (receiver is a scene type post package-split).
func initDerived(p *Project, rifxRoot *rifx.Chunk) {
	var maxID uint32
	for _, c := range p.Compositions {
		if c.ID > maxID {
			maxID = c.ID
		}
		for _, l := range c.Layers {
			if l.ID > maxID {
				maxID = l.ID
			}
		}
		// Comps also carry non-parsed service layers (DLay/SLay/CLay/SecL —
		// present in AE files and in our dummy-comp template with IDs 2..12)
		// whose ldta IDs live in the same head-counter namespace but never
		// enter c.Layers. Scan the raw itemList so allocItemID can't collide
		// with them (AE 2025 rejects such collisions with "unexpected match
		// name searched for in group"; see nextitemid-must-include-layer-ids).
		if cb := compositionBack(c); cb != nil && cb.itemList != nil {
			if m := maxLayerIDInItemList(cb.itemList); m > maxID {
				maxID = m
			}
		}
	}
	for _, f := range p.Footage {
		if f.ID > maxID {
			maxID = f.ID
		}
	}
	for _, fo := range p.Folders {
		if fo.ID > maxID {
			maxID = fo.ID
		}
	}
	scene.SetProjectNextItemID(p, maxID+1)

	pb := projectBack(p)
	if pb == nil {
		pb = &projectBackrefs{}
		scene.SetProjectBack(p, pb)
	}
	for _, c := range rifxRoot.Children {
		if c.IsList() && c.FormType == rifx.IDFold {
			pb.rootFold = c
			break
		}
	}
}

// allocItemID 返回下一个可用 Item ID 并递增计数器。Monotonic，不 reuse。
// Free function (receiver is a scene type post package-split).
func allocItemID(p *Project) uint32 {
	id := scene.ProjectNextItemID(p)
	scene.SetProjectNextItemID(p, id+1)
	return id
}

// parseItem classifies an Item list and dispatches to the right handler.
func parseItem(item *rifx.Chunk, proj *Project) error {
	itemType, itemID, err := classifyItem(item)
	if err != nil {
		return err
	}

	name := ""
	if utf8 := item.FindFirst(rifx.IDUtf8); utf8 != nil {
		name = utf8.Text()
	}

	// Item-level metadata shared across all item types.
	cmta := item.FindFirst(rifx.IDCmta)
	idta := item.FindFirst(rifx.IDIdta)
	var comment string
	if cmta != nil {
		comment = decodeCmta(cmta.Data)
	}
	var label uint8
	if idta != nil && len(idta.Data) > 0x3A {
		label = idta.Data[0x3A]
	}

	switch itemType {
	case ItemTypeComposition:
		comp, err := parseCompositionIntoProject(item, itemID, name, proj)
		if err != nil {
			return fmt.Errorf("aep: parse comp %q: %w", name, err)
		}
		scene.SetCompositionProj(comp, proj) // wire back-pointer so Layer.SourceComposition() works
		comp.Comment = comment
		comp.Label = label
		cb := compositionBack(comp)
		if cb == nil {
			cb = &compositionBackrefs{compName: comp.Name}
			scene.SetCompositionBack(comp, cb)
		}
		cb.itemCmtaChunk = cmta
		cb.itemIdtaChunk = idta
		cb.itemLayrParent = item
		proj.Compositions = append(proj.Compositions, comp)

	case ItemTypeFootage:
		footage, err := parseFootage(item, itemID, name)
		if err != nil {
			return fmt.Errorf("aep: parse footage %q: %w", name, err)
		}
		footage.Comment = comment
		footage.Label = label
		fb := footageBack(footage)
		if fb == nil {
			fb = &footageBackrefs{itemID: footage.ID, itemName: footage.Name}
			scene.SetFootageBack(footage, fb)
		}
		fb.itemID = footage.ID
		fb.itemName = footage.Name
		fb.itemCmtaChunk = cmta
		fb.itemIdtaChunk = idta
		fb.itemLayrParent = item
		proj.Footage = append(proj.Footage, footage)

	case ItemTypeFolder:
		proj.Folders = append(proj.Folders, &Folder{ID: itemID, Name: name})
	}

	return nil
}

// classifyItem returns the item type and ID from an Item list.
//
// idta layout per boltframe/aftereffects-aep-parser IDTA struct:
//
//	0x00–0x01 : Type    (uint16 BE)  1=folder, 4=composition, 7=footage
//	0x02–0x0F : unknown / reserved
//	0x10–0x13 : ID      (uint32 BE)
func classifyItem(item *rifx.Chunk) (ItemType, uint32, error) {
	idta := item.FindFirst(rifx.IDIdta)
	if idta == nil || len(idta.Data) < 20 {
		return ItemTypeUnknown, 0, nil
	}

	typeRaw, err := idta.U16(0)
	if err != nil {
		return ItemTypeUnknown, 0, err
	}
	id, err := idta.U32(16)
	if err != nil {
		return ItemTypeUnknown, 0, err
	}

	var t ItemType
	switch typeRaw {
	case 0x01:
		t = ItemTypeFolder
	case 0x04:
		t = ItemTypeComposition
	case 0x07:
		t = ItemTypeFootage
	default:
		t = ItemTypeUnknown
	}
	return t, id, nil
}

// walkTdmnPairs iterates a tdgp-style group's children as tdmn + payload
// pairs. fn is called with each named entry's name and its payload LIST
// (the next sibling). Returning false from fn stops iteration. The "ADBE
// Group End" sentinel terminates automatically. Non-LIST payloads and
// orphan tdmn entries are skipped.
//
// Used for any chunk laid out as alternating tdmn + payload pairs
// (property groups, mask atom bodies, effect parade entries, etc.).
// Mask atoms — where tdmn is followed by a non-LIST mkif chunk before
// the tdgp — are walked by hand in parse_mask.go.
func walkTdmnPairs(group *rifx.Chunk, fn func(name string, payload *rifx.Chunk) bool) {
	kids := group.Children
	for i := 0; i < len(kids); i++ {
		ch := kids[i]
		if ch.ID != rifx.IDTdmn {
			continue
		}
		name := trimNUL(ch.Data)
		if name == "ADBE Group End" {
			return
		}
		if i+1 >= len(kids) {
			continue
		}
		payload := kids[i+1]
		if !payload.IsList() {
			continue
		}
		i++
		if !fn(name, payload) {
			return
		}
	}
}

// trimNUL returns s up to the first NUL byte (AE strings are NUL-padded
// to a fixed footprint).
func trimNUL(b []byte) string {
	end := len(b)
	for end > 0 && b[end-1] == 0 {
		end--
	}
	return string(b[:end])
}

// readFloat64BE decodes a big-endian IEEE 754 double-precision float from
// b at offset, reporting false if the bounds don't fit.
func readFloat64BE(b []byte, offset int) (float64, bool) {
	if offset < 0 || offset+8 > len(b) {
		return 0, false
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b[offset:])), true
}

// readFloat64LE decodes a little-endian IEEE 754 double from b at offset.
// Only the otst orientation cdat stores values little-endian (see
// parseOrientationProperty); everything else is big-endian.
func readFloat64LE(b []byte, offset int) (float64, bool) {
	if offset < 0 || offset+8 > len(b) {
		return 0, false
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b[offset:])), true
}
