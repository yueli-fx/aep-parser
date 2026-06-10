package serializer

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"path/filepath"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// parseFootage reads footage metadata from an Item list.
//
// Real .aep files wrap footage descriptors in a "Pin " sublist holding:
//   - sspc  : width/height (+ framerate in extended versions)
//   - opti  : type tag ("png!", "ZPEG", "Soli", "Plac", ...) and sometimes a name
//   - Als2/alas : JSON alias data with "fullpath" — most reliable source of name
func parseFootage(item *rifx.Chunk, id uint32, fallbackName string) (*Footage, error) {
	fb := &footageBackrefs{itemID: id, itemName: fallbackName}
	footage := &Footage{ID: id, Name: fallbackName}
	scene.SetFootageBack(footage, fb)

	pin := item.FindFirstList(rifx.IDPin)
	src := item
	if pin != nil {
		src = pin
	}

	if cpth := src.FindFirst(rifx.IDCpth); cpth != nil {
		footage.Path = cpth.Text()
		fb.cpthChunk = cpth
	}

	if sspc := src.FindFirst(rifx.IDSspc); sspc != nil && len(sspc.Data) >= 4 {
		fb.sspcChunk = sspc
		// Real AE sspc is 222 bytes with width/height at @0x20/@0x24
		// (py-aep binary/footage_chunks.py::SspcChunk). Synthesized test
		// fixtures use a short 4-byte sspc with width/height at byte 0/2.
		// Pick the layout based on chunk length.
		if len(sspc.Data) >= 38 {
			footage.Width = binary.BigEndian.Uint16(sspc.Data[0x20:0x22])
			footage.Height = binary.BigEndian.Uint16(sspc.Data[0x24:0x26])
		} else {
			footage.Width = binary.BigEndian.Uint16(sspc.Data[0:2])
			footage.Height = binary.BigEndian.Uint16(sspc.Data[2:4])
		}
	}

	if opti := src.FindFirst(rifx.IDOpti); opti != nil {
		fb.optiChunk = opti
		kind, name := parseOpti(opti.Data)
		if name != "" && footage.Name == "" {
			footage.Name = name
		}
		if kind == "Soli" {
			footage.IsSolid = true
			// ARGB 4×float32 BE at @0x0A (RE: re_solidnull.aep); alpha is
			// always 1.0 in AE-written solids and is not surfaced.
			if len(opti.Data) >= optiSoliColorB+4 {
				footage.SolidColor = [3]float64{
					float64(math.Float32frombits(binary.BigEndian.Uint32(opti.Data[optiSoliColorR : optiSoliColorR+4]))),
					float64(math.Float32frombits(binary.BigEndian.Uint32(opti.Data[optiSoliColorG : optiSoliColorG+4]))),
					float64(math.Float32frombits(binary.BigEndian.Uint32(opti.Data[optiSoliColorB : optiSoliColorB+4]))),
				}
			}
		}
		if kind == "Plac" {
			footage.IsPlaceholder = true
		}
	}

	// Look for Als2/alas JSON anywhere in this Item subtree.
	if alas, path, base := findAliasPath(item); alas != nil {
		fb.aliasChunk = alas
		footage.Path = path
		if base != "" {
			footage.Name = base
		}
	}

	// If we still have nothing and no path, treat as solid (boltframe convention).
	if footage.Name == "" && footage.Path == "" {
		footage.IsSolid = true
	}

	return footage, nil
}

// parseOpti decodes the footage-options chunk. The kind tag is 4 bytes at
// offset 0 (case- and order-significant). Names live at kind-specific offsets:
//
//	"Soli"      solid color        — name at 0x1A (boltframe)
//	"Plac"      placeholder        — name at 0x0A (boltframe)
//	"png!", "ZPEG", ... (file)     — name/ext at 0x3A; the real filename is
//	                                 usually in the sibling Als2/alas chunk
func parseOpti(d []byte) (kind, name string) {
	if len(d) < 4 {
		return "", ""
	}
	kind = string(d[:4])
	start := 0
	switch kind {
	case "Plac":
		start = 0x0A
	case "Soli":
		start = 0x1A
	default:
		start = 0x3A
	}
	if start >= len(d) {
		return kind, ""
	}
	end := start
	for end < len(d) && d[end] != 0 {
		end++
	}
	return kind, strings.TrimSpace(string(d[start:end]))
}

// findAliasPath walks an Item subtree looking for an alas JSON blob with a
// "fullpath" field. Returns (chunk, fullpath, basename) — chunk is nil when
// no alas was found.
func findAliasPath(item *rifx.Chunk) (chunk *rifx.Chunk, fullpath, base string) {
	var visit func(c *rifx.Chunk)
	visit = func(c *rifx.Chunk) {
		if chunk != nil {
			return
		}
		if c.ID == rifx.IDAlas && len(c.Data) > 0 && c.Data[0] == '{' {
			var alias struct {
				FullPath string `json:"fullpath"`
			}
			if err := json.Unmarshal(c.Data, &alias); err == nil && alias.FullPath != "" {
				chunk = c
				fullpath = alias.FullPath
				base = filepath.Base(strings.ReplaceAll(fullpath, `\`, `/`))
			}
			return
		}
		for _, ch := range c.Children {
			visit(ch)
		}
	}
	visit(item)
	return chunk, fullpath, base
}
