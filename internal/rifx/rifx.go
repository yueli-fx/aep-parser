// Package rifx implements a Big-Endian RIFF (RIFX) binary parser.
// Adobe After Effects .aep files use the RIFX format: magic "RIFX" + form-type "Egg!".
package rifx

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ChunkID is a 4-byte ASCII chunk identifier.
type ChunkID [4]byte

func (c ChunkID) String() string { return string(c[:]) }

// Well-known chunk IDs used in .aep files.
var (
	IDList = ChunkID{'L', 'I', 'S', 'T'}
	IDRifx = ChunkID{'R', 'I', 'F', 'X'}

	IDEgg  = ChunkID{'E', 'g', 'g', '!'} // root form type
	IDItem = ChunkID{'I', 't', 'e', 'm'} // project item
	IDLayr = ChunkID{'L', 'a', 'y', 'r'} // layer

	IDUtf8 = ChunkID{'U', 't', 'f', '8'} // UTF-8 name
	IDCdta = ChunkID{'c', 'd', 't', 'a'} // composition data
	IDIdta = ChunkID{'i', 'd', 't', 'a'} // item metadata
	IDLdta = ChunkID{'l', 'd', 't', 'a'} // layer data
	IDSspc = ChunkID{'s', 's', 'p', 'c'} // source/footage reference
	IDFnam = ChunkID{'f', 'n', 'a', 'm'} // effect instance name (in a project: embeds a "Utf8" sub-record; in a .ffx preset: fixed-width raw)
	IDFdta = ChunkID{'f', 'd', 't', 'a'} // folder marker
	IDCmta = ChunkID{'c', 'm', 't', 'a'} // comment
	IDKfmd = ChunkID{'k', 'f', 'm', 'd'} // keyframe data
	IDStbl = ChunkID{'s', 't', 'b', 'l'} // property stream table
	IDCpth = ChunkID{'C', 'p', 't', 'h'} // footage path
	IDOpti = ChunkID{'o', 'p', 't', 'i'} // footage options (contains name)
	IDPin  = ChunkID{'P', 'i', 'n', ' '} // pin container (wraps footage descriptors)
	IDSfdr = ChunkID{'S', 'f', 'd', 'r'} // sub-folder records
	IDFold = ChunkID{'F', 'o', 'l', 'd'} // project root folder container
	IDAls2 = ChunkID{'A', 'l', 's', '2'} // alias-2 wrapper (contains alas)
	IDAlas = ChunkID{'a', 'l', 'a', 's'} // alias data (JSON with fullpath in newer AE)
	IDTdb4 = ChunkID{'T', 'd', 'b', '4'} // legacy uppercase tdb4 (older AE)
	IDtdb4 = ChunkID{'t', 'd', 'b', '4'} // property descriptor (124 bytes)
	IDTdmn = ChunkID{'t', 'd', 'm', 'n'} // property match name (NUL-padded)
	IDTdgp = ChunkID{'t', 'd', 'g', 'p'} // property group LIST formType
	IDTdbs = ChunkID{'t', 'd', 'b', 's'} // property block LIST formType
	IDCdat = ChunkID{'c', 'd', 'a', 't'} // current/static property value
	IDLhd3 = ChunkID{'l', 'h', 'd', '3'} // keyframe-list header (52 bytes)
	IDLdat = ChunkID{'l', 'd', 'a', 't'} // keyframe stream data
	IDkfl  = ChunkID{'l', 'i', 's', 't'} // LIST formType for keyframe list container
	IDMrst = ChunkID{'m', 'r', 's', 't'} // marker/keyframe set
	IDBtds = ChunkID{'b', 't', 'd', 's'} // text-layer source data (opaque LIST)
	IDBtdk = ChunkID{'b', 't', 'd', 'k'} // text-layer source key (opaque LIST inside btds)
	IDMrky = ChunkID{'m', 'r', 'k', 'y'} // marker keys container LIST (one Nmrd per marker)
	IDNmrd = ChunkID{'N', 'm', 'r', 'd'} // numbered marker data LIST (one marker's metadata)
	IDNmhd = ChunkID{'N', 'm', 'H', 'd'} // marker NmHd header (17-20 bytes: flags, duration, label)
	IDOmks = ChunkID{'o', 'm', 'k', 's'} // mask shape container LIST
	IDOmS  = ChunkID{'o', 'm', '-', 's'} // mask shape inner wrapper LIST
	IDShap = ChunkID{'s', 'h', 'a', 'p'} // shape (path) data LIST
	IDShph = ChunkID{'s', 'h', 'p', 'h'} // shape header (24 bytes; closed flag at 0x14)
	IDOmtn = ChunkID{'o', 'm', 't', 'n'} // mask/shape title/name (Utf8-like)
	IDMkif = ChunkID{'m', 'k', 'i', 'f'} // mask info (48 bytes; mode, inverted, color)
	IDSecL = ChunkID{'S', 'e', 'c', 'L'} // section-layer LIST (pseudo-layer holding comp-level props like "Markers")
	IDTdsb = ChunkID{'t', 'd', 's', 'b'} // property subprop flags (4 bytes: byte 2 bit 4 = locked_ratio, etc.)
	IDtdum = ChunkID{'t', 'd', 'u', 'm'} // property min value (variable: f32×4 color | u32 integer | f64×N)
	IDtduM = ChunkID{'t', 'd', 'u', 'M'} // property max value (same layout as tdum)
	IDTdsn = ChunkID{'t', 'd', 's', 'n'} // property / group display-name carrier (embeds a "Utf8" sub-record in its payload)
	IDNhed = ChunkID{'n', 'h', 'e', 'd'} // project header (32-byte payload, holds BitsPerChannel @0x0F)
	IDNnhd = ChunkID{'n', 'n', 'h', 'd'} // project secondary header (40-byte payload, BitsPerChannel mirror @0x18)
	IDOvG2 = ChunkID{'O', 'v', 'G', '2'} // Essential Properties override container LIST (sibling of tdmn "ADBE Layer Overrides")
	IDBlsv = ChunkID{'b', 'l', 's', 'v'} // Block "Layer Source Version" — 4-byte BE uint, seen value 1
	IDBlsi = ChunkID{'b', 'l', 's', 'i'} // Block "Layer Source Item id" — 4-byte BE uint (AVItem id; 0 = no override)
	IDPRin = ChunkID{'P', 'R', 'i', 'n'} // LIST formType: comp's pre-render-info container (renderer + renderer-data)
	IDPrin = ChunkID{'p', 'r', 'i', 'n'} // pre-render-info chunk (104 bytes; renderer name as ASCII @offset 4)
	IDPrda = ChunkID{'p', 'r', 'd', 'a'} // pre-render data chunk (renderer-specific options; variable length)
	IDOtst = ChunkID{'o', 't', 's', 't'} // LIST formType: orientation wrapper (holds tdbs + otky)
	IDOtky = ChunkID{'o', 't', 'k', 'y'} // LIST formType: orientation-keyframe container (holds otda)
	IDOtda = ChunkID{'o', 't', 'd', 'a'} // orientation default value chunk (24 B = 3 × f64)
	IDGide = ChunkID{'G', 'i', 'd', 'e'} // LIST formType: layer-side guide chunk (4th Layr child; AE-required boilerplate per iter-5 RE)
	IDGdta = ChunkID{'g', 'd', 't', 'a'} // chunk inside Gide (8 B all zero observed)
	IDEwst = ChunkID{'E', 'w', 's', 't'} // LIST formType: empty 0-child sibling of every Layr at Item level (AE-required boilerplate per iter-5 RE)
	IDTdpi = ChunkID{'t', 'd', 'p', 'i'} // property host-layer binding (4-byte BE layer id inside an effect param's tdbs; AE validates it resolves on open)
	IDTdps = ChunkID{'t', 'd', 'p', 's'} // companion of tdpi (4-byte, observed 0) in a layer-picker value entry
	IDPefl = ChunkID{'P', 'e', 'f', 'l'} // LIST formType: project-level effect-list container (children = pjef Utf8 entries naming used effects)
	IDPjef = ChunkID{'p', 'j', 'e', 'f'} // Utf8 chunk: one effect match-name inside Pefl
	IDparT = ChunkID{'p', 'a', 'r', 'T'} // LIST formType: effect parameter definitions container (pard entries)
	IDpard = ChunkID{'p', 'a', 'r', 'd'} // effect parameter definition chunk (variable-length, keyed by control type)
	IDParn = ChunkID{'p', 'a', 'r', 'n'} // parT param-count header (u32 = number of pard entries, incl. ADBE Effect Built In Params)
	IDPgui = ChunkID{'p', 'g', 'u', 'i'} // effect sspc trailing UI-state chunk (16 bytes, group fold/expand flags)
	IDPdnm = ChunkID{'p', 'd', 'n', 'm'} // param display-name / aux string (Utf8): checkbox on-off label, dropdown "opt1|opt2" options
	IDBesc = ChunkID{'b', 'e', 's', 'c'} // LIST formType: animation-preset (.ffx) descriptor wrapper (root child of a RIFX "FaFX" form)
	IDTdsp = ChunkID{'t', 'd', 's', 'p'} // LIST formType: preset target property-path descriptor (holds tdsi path steps)
	IDTdsi = ChunkID{'t', 'd', 's', 'i'} // LIST formType: one property-path step (tdix index + tdmn match-name)
	IDGCst = ChunkID{'G', 'C', 's', 't'} // LIST formType: gradient color stops wrapper (contains tdbs + GCky)
	IDGCky = ChunkID{'G', 'C', 'k', 'y'} // LIST formType: gradient keyframe container (Utf8 children holding prop.map XML, one per keyframe)

	// Essential Graphics panel. Item-level CIF3.
	IDCif3 = ChunkID{'C', 'I', 'F', '3'} // LIST formType: EG panel definition (template-name CpS2 + CCtl controllers)
	IDCifO = ChunkID{'C', 'I', 'F', 'O'} // LIST formType: oldest EG panel generation; AE writes CIFO+CIF2+CIF3 with byte-identical content
	IDCif2 = ChunkID{'C', 'I', 'F', '2'} // LIST formType: middle EG panel generation (see IDCifO)
	IDCapS = ChunkID{'C', 'a', 'p', 'S'} // LIST formType: EG caption string (CsCt + CapL + Utf8 value, no locale — sibling of CpS2)
	IDCsCt = ChunkID{'C', 's', 'C', 't'} // U4 little-endian: EG string count inside CpS2/CapS (observed 1)
	IDCapL = ChunkID{'C', 'a', 'p', 'L'} // U4: EG caption locale selector (observed 0)
	IDCcCt = ChunkID{'C', 'c', 'C', 't'} // U4 big-endian: EG controller count per CIF* generation
	IDCprC = ChunkID{'C', 'p', 'r', 'C'} // U4: CPrp count — big-endian inside CCtl, little-endian inside OvG2 (observed)
	IDCPrp = ChunkID{'C', 'P', 'r', 'p'} // LIST formType: EG property ref — in CCtl: CCId+CLId+Utf8 JSON path; in OvG2: bare Utf8 uuid
	IDCCId = ChunkID{'C', 'C', 'I', 'd'} // U4 big-endian: EG ref comp item ID
	IDCLId = ChunkID{'C', 'L', 'I', 'd'} // U4 big-endian: EG ref host layer ID
	IDCVal = ChunkID{'C', 'V', 'a', 'l'} // EG controller current value (layout per CTyp: slider f64, checkbox u1, color 4xf32 RGBA, point 2xf64)
	IDCDef = ChunkID{'C', 'D', 'e', 'f'} // EG controller default value (same layout as CVal)
	IDSmin = ChunkID{'S', 'm', 'i', 'n'} // F8 big-endian: EG slider minimum
	IDSmax = ChunkID{'S', 'm', 'a', 'x'} // F8 big-endian: EG slider maximum
	IDCctl = ChunkID{'C', 'C', 't', 'l'} // LIST formType: one EG controller (CpS2 name + Utf8 uuid + CTyp type)
	IDCpS2 = ChunkID{'C', 'p', 'S', '2'} // LIST formType: localized string (Utf8 value + locale)
	IDCTyp = ChunkID{'C', 'T', 'y', 'p'} // U4: EG controller type (1=Checkbox 2=Slider 4=Color 5=Point 6=Text 8=Comment 9=MultiDim 10=Group 13=Dropdown)

	// Project-level setting chunks
	IDAcer = ChunkID{'a', 'c', 'e', 'r'} // U1: compensate_for_scene_referred_profiles
	IDAdfr = ChunkID{'a', 'd', 'f', 'r'} // F8: audio sample rate (Hz)
	IDDwga = ChunkID{'d', 'w', 'g', 'a'} // Variable (1-4B observed): byte 0 = working gamma selector (0 → 2.2, ≠0 → 2.4)
	IDLnrb = ChunkID{'l', 'n', 'r', 'b'} // flag chunk: linear_blending (presence = true)
	IDLnrp = ChunkID{'l', 'n', 'r', 'p'} // flag chunk: linearize_working_space (presence = true)
	IDGpuG = ChunkID{'g', 'p', 'u', 'G'} // LIST formType: GPU device id container (single Utf8 child = UUID)
	IDExEn = ChunkID{'E', 'x', 'E', 'n'} // LIST formType: expression engine container (single Utf8 child = "extendscript" / "javascript-1.0")

	// Render queue chunk family. IDAls2/IDAlas/IDLhd3/IDLdat already above; "list" = IDkfl.
	IDLRdr = ChunkID{'L', 'R', 'd', 'r'} // LIST formType: render queue container (direct root child)
	IDLItm = ChunkID{'L', 'I', 't', 'm'} // LIST formType: render-queue item collection (per-item RCom/list/LOm groups)
	IDLOm  = ChunkID{'L', 'O', 'm', ' '} // LIST formType: output-module group (Roou-delimited modules + Als2 + Utf8 name/template)
	IDRCom = ChunkID{'R', 'C', 'o', 'm'} // non-LIST container leaf: render-queue item comment (embeds a Utf8 chunk in Data)
	IDRoou = ChunkID{'R', 'o', 'o', 'u'} // leaf: output-module settings (154+ bytes); delimits output modules within LOm
	IDRopt = ChunkID{'R', 'o', 'p', 't'} // leaf: format-specific render options (polymorphic by format_code)
	IDRout = ChunkID{'R', 'o', 'u', 't'} // leaf: render-queue item flags (4B header + 4B/item)
)

// opaqueListTypes is the set of LIST formTypes whose payload is non-chunk
// binary data (e.g. serialized text-layer source). These LISTs are read as
// raw bytes into Chunk.Data and written back verbatim — no recursion.
var opaqueListTypes = map[ChunkID]bool{
	IDBtds: true,
	IDBtdk: true,
}

// Chunk is a parsed RIFX node.
type Chunk struct {
	ID       ChunkID
	Size     uint32
	Data     []byte   // leaf chunks only
	FormType ChunkID  // LIST/RIFX chunks only
	Children []*Chunk // LIST/RIFX chunks only

	// Trailing holds bytes that follow this chunk's regular payload but are
	// not represented in the structured fields above. For the root chunk
	// returned by Parse, this captures any data after the RIFX chunk's
	// declared size — real .aep files often carry a few KB of opaque tail
	// data here. Trailing is written verbatim by Write after the chunk slot
	// and is *not* counted in PayloadSize.
	Trailing []byte
}

// IsList reports whether this is a container chunk.
func (c *Chunk) IsList() bool { return c.ID == IDList || c.ID == IDRifx }

// FindFirst returns the first direct child with the given ID.
func (c *Chunk) FindFirst(id ChunkID) *Chunk {
	for _, ch := range c.Children {
		if ch.ID == id {
			return ch
		}
	}
	return nil
}

// FindAll returns all direct children with the given ID.
func (c *Chunk) FindAll(id ChunkID) []*Chunk {
	var out []*Chunk
	for _, ch := range c.Children {
		if ch.ID == id {
			out = append(out, ch)
		}
	}
	return out
}

// FindFirstList returns the first LIST child matching the given FormType.
func (c *Chunk) FindFirstList(ft ChunkID) *Chunk {
	for _, ch := range c.Children {
		if ch.IsList() && ch.FormType == ft {
			return ch
		}
	}
	return nil
}

// FindAllList returns all LIST children matching the given FormType.
func (c *Chunk) FindAllList(ft ChunkID) []*Chunk {
	var out []*Chunk
	for _, ch := range c.Children {
		if ch.IsList() && ch.FormType == ft {
			out = append(out, ch)
		}
	}
	return out
}

// U8 reads one byte from Data at offset.
func (c *Chunk) U8(offset int) (byte, error) {
	if offset < 0 || offset >= len(c.Data) {
		return 0, fmt.Errorf("rifx: U8 offset %d OOB (len=%d)", offset, len(c.Data))
	}
	return c.Data[offset], nil
}

// U16 reads a big-endian uint16 from Data at offset.
func (c *Chunk) U16(offset int) (uint16, error) {
	if offset < 0 || offset > len(c.Data)-2 {
		return 0, fmt.Errorf("rifx: U16 offset %d OOB (len=%d)", offset, len(c.Data))
	}
	return binary.BigEndian.Uint16(c.Data[offset:]), nil
}

// U32 reads a big-endian uint32 from Data at offset.
func (c *Chunk) U32(offset int) (uint32, error) {
	if offset < 0 || offset > len(c.Data)-4 {
		return 0, fmt.Errorf("rifx: U32 offset %d OOB (len=%d)", offset, len(c.Data))
	}
	return binary.BigEndian.Uint32(c.Data[offset:]), nil
}

// Text returns chunk Data as a UTF-8 string (null bytes stripped).
func (c *Chunk) Text() string {
	b := c.Data
	for len(b) > 0 && b[len(b)-1] == 0 {
		b = b[:len(b)-1]
	}
	return string(b)
}

// IsOpaqueList reports whether this LIST chunk holds non-chunk binary in
// Data (instead of recursive Children). True for LIST types in
// opaqueListTypes (e.g. btds, btdk).
func (c *Chunk) IsOpaqueList() bool {
	return c.IsList() && opaqueListTypes[c.FormType]
}

// PayloadSize returns the size that should be written in this chunk's size
// header. For leaf chunks it is len(Data); for opaque LIST chunks it is 4
// (formType) + len(Data); for structured LIST chunks it is 4 (formType)
// plus the sum of each child's full slot (8-byte header + payload + 1 byte
// pad when payload is odd). The stored Size field is ignored — sizes are
// always recomputed so callers may freely mutate Data and Children.
func (c *Chunk) PayloadSize() uint32 {
	if c.IsOpaqueList() {
		return 4 + uint32(len(c.Data))
	}
	if c.IsList() {
		sz := uint32(4) // FormType
		for _, ch := range c.Children {
			child := ch.PayloadSize()
			sz += 8 + child
			if child%2 != 0 {
				sz++ // pad byte
			}
		}
		return sz
	}
	return uint32(len(c.Data))
}

// Write serializes the chunk tree to w in the same RIFX byte layout that
// Parse reads. Sizes are recomputed from current Data/Children — see
// PayloadSize. Trailing bytes (if any) are appended verbatim after the
// chunk slot.
func (c *Chunk) Write(w io.Writer) error {
	size := c.PayloadSize()
	if _, err := w.Write(c.ID[:]); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, size); err != nil {
		return err
	}
	if c.IsList() {
		if _, err := w.Write(c.FormType[:]); err != nil {
			return err
		}
		if c.IsOpaqueList() {
			if _, err := w.Write(c.Data); err != nil {
				return err
			}
			if size%2 != 0 {
				if _, err := w.Write([]byte{0}); err != nil {
					return err
				}
			}
		} else {
			for _, ch := range c.Children {
				if err := ch.Write(w); err != nil {
					return err
				}
			}
		}
	} else {
		if _, err := w.Write(c.Data); err != nil {
			return err
		}
		if size%2 != 0 {
			if _, err := w.Write([]byte{0}); err != nil {
				return err
			}
		}
	}
	if len(c.Trailing) > 0 {
		if _, err := w.Write(c.Trailing); err != nil {
			return err
		}
	}
	return nil
}

// Parse reads a RIFX file from r and returns the root chunk. Any bytes
// after the RIFX chunk's declared size are captured into root.Trailing so
// that Write can re-emit them — real .aep files carry opaque tail data
// there that AE expects to find.
func Parse(r io.ReadSeeker) (*Chunk, error) {
	return ParseWithLimits(r, DefaultLimits)
}

// ParseWithLimits reads a RIFX file while enforcing explicit resource limits.
func ParseWithLimits(r io.ReadSeeker, limits Limits) (*Chunk, error) {
	p, err := newParser(r, limits)
	if err != nil {
		return nil, err
	}
	root, err := p.readChunk(0, p.inputEnd)
	if err != nil {
		return nil, fmt.Errorf("rifx: read root: %w", err)
	}
	if root.ID != IDRifx {
		return nil, fmt.Errorf("rifx: not a RIFX file (got magic %q)", root.ID)
	}
	if root.FormType != IDEgg {
		return nil, fmt.Errorf("rifx: unexpected form type %q (want \"Egg!\")", root.FormType)
	}
	trailing, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("rifx: read trailing: %w", err)
	}
	if len(trailing) > 0 {
		root.Trailing = trailing
	}
	return root, nil
}

// ReadChunk parses a single chunk (header + body) from r, recursing into
// LIST containers. Unlike Parse, the input is not required to be a RIFX
// root — useful for embedded resource blobs that store a single LIST chunk.
func ReadChunk(r io.ReadSeeker) (*Chunk, error) { return ReadChunkWithLimits(r, DefaultLimits) }

// ReadChunkWithLimits parses one chunk while enforcing explicit resource limits.
func ReadChunkWithLimits(r io.ReadSeeker, limits Limits) (*Chunk, error) {
	p, err := newParser(r, limits)
	if err != nil {
		return nil, err
	}
	return p.readChunk(0, p.inputEnd)
}
