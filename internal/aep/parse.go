package aep

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

// Open parses an .aep file by path and returns the Project.
func Open(path string) (*Project, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("aep: open %q: %w", path, err)
	}
	defer f.Close()
	return FromReader(f)
}

// FromReader parses an .aep file from an io.ReadSeeker.
func FromReader(r io.ReadSeeker) (*Project, error) {
	root, err := rifx.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("aep: parse RIFX: %w", err)
	}
	proj, err := parseProject(root)
	if err != nil {
		return nil, err
	}
	proj.root = root
	return proj, nil
}

func parseProject(root *rifx.Chunk) (*Project, error) {
	proj := &Project{}

	// Project-level header chunks: nhed (32-byte) + nnhd (40-byte) sit
	// as direct children of the root LIST/RIFX. Both carry a copy of
	// BitsPerChannel at the same byte position relative to chunk start.
	if nhed := root.FindFirst(rifx.IDNhed); nhed != nil {
		proj.nhedChunk = nhed
		if len(nhed.Data) > 0x0F {
			proj.BitsPerChannel = BitsPerChannel(nhed.Data[0x0F])
		}
	}
	if nnhd := root.FindFirst(rifx.IDNnhd); nnhd != nil {
		proj.nnhdChunk = nnhd
	}

	// In real .aep files, Item lists are nested inside Fold/Sfdr containers,
	// not direct children of the root. Walk the whole tree.
	var walk func(c *rifx.Chunk) error
	walk = func(c *rifx.Chunk) error {
		for _, child := range c.Children {
			if !child.IsList() {
				continue
			}
			if child.FormType == rifx.IDItem {
				if err := parseItem(child, proj); err != nil {
					return err
				}
			}
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root); err != nil {
		return nil, err
	}
	proj.initDerived(root)
	return proj, nil
}

// initDerived 在 parseProject 收尾时调用，初始化 Project 的 derived state：
//   - nextItemID = max(已有所有 item IDs) + 1（monotonic counter for NewComposition）
//   - rootFold = root Egg! 下第一个 formType=Fold 的 LIST（cached for V2 mutations）
//
// 参数 rifxRoot 是 parseProject 顶层 *rifx.Chunk（formType=Egg!）。
func (p *Project) initDerived(rifxRoot *rifx.Chunk) {
	var maxID uint32
	for _, c := range p.Compositions {
		if c.ID > maxID {
			maxID = c.ID
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
	p.nextItemID = maxID + 1

	for _, c := range rifxRoot.Children {
		if c.IsList() && c.FormType == rifx.IDFold {
			p.rootFold = c
			break
		}
	}
}

// allocItemID 返回下一个可用 Item ID 并递增计数器。Monotonic，不 reuse（见 Invariant #9）。
func (p *Project) allocItemID() uint32 {
	id := p.nextItemID
	p.nextItemID++
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
		comp, err := parseComposition(item, itemID, name, &proj.Warnings)
		if err != nil {
			return fmt.Errorf("aep: parse comp %q: %w", name, err)
		}
		comp.proj = proj // wire back-pointer so Layer.SourceComposition() works
		comp.Comment = comment
		comp.Label = label
		comp.itemCmtaChunk = cmta
		comp.itemIdtaChunk = idta
		comp.itemLayrParent = item
		proj.Compositions = append(proj.Compositions, comp)

	case ItemTypeFootage:
		footage, err := parseFootage(item, itemID, name)
		if err != nil {
			return fmt.Errorf("aep: parse footage %q: %w", name, err)
		}
		footage.Comment = comment
		footage.Label = label
		footage.itemCmtaChunk = cmta
		footage.itemIdtaChunk = idta
		footage.itemLayrParent = item
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
