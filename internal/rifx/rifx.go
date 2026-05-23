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
	if offset >= len(c.Data) {
		return 0, fmt.Errorf("rifx: U8 offset %d OOB (len=%d)", offset, len(c.Data))
	}
	return c.Data[offset], nil
}

// U16 reads a big-endian uint16 from Data at offset.
func (c *Chunk) U16(offset int) (uint16, error) {
	if offset+2 > len(c.Data) {
		return 0, fmt.Errorf("rifx: U16 offset %d OOB (len=%d)", offset, len(c.Data))
	}
	return binary.BigEndian.Uint16(c.Data[offset:]), nil
}

// U32 reads a big-endian uint32 from Data at offset.
func (c *Chunk) U32(offset int) (uint32, error) {
	if offset+4 > len(c.Data) {
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
	root, err := readChunk(r)
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

func readChunk(r io.ReadSeeker) (*Chunk, error) {
	var id ChunkID
	if _, err := io.ReadFull(r, id[:]); err != nil {
		return nil, fmt.Errorf("read chunk ID: %w", err)
	}

	var size uint32
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, fmt.Errorf("chunk %q: read size: %w", id, err)
	}

	chunk := &Chunk{ID: id, Size: size}

	if id == IDList || id == IDRifx {
		if _, err := io.ReadFull(r, chunk.FormType[:]); err != nil {
			return nil, fmt.Errorf("chunk %q: read form type: %w", id, err)
		}
		// Opaque LIST types (e.g. text-source btds/btdk) hold non-chunk
		// binary — read entire payload as Data, no recursion.
		if opaqueListTypes[chunk.FormType] {
			data := make([]byte, int(size)-4)
			if _, err := io.ReadFull(r, data); err != nil {
				return nil, fmt.Errorf("chunk %q/%q: read opaque payload (size=%d): %w", id, chunk.FormType, size, err)
			}
			chunk.Data = data
			if size%2 != 0 {
				if _, err := r.Seek(1, io.SeekCurrent); err != nil {
					return nil, err
				}
			}
			return chunk, nil
		}
		start, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		end := start + int64(size) - 4
		for {
			pos, err := r.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil, err
			}
			if pos >= end {
				break
			}
			child, err := readChunk(r)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					break
				}
				return nil, fmt.Errorf("child of %q/%q: %w", id, chunk.FormType, err)
			}
			chunk.Children = append(chunk.Children, child)
		}
		if _, err := r.Seek(end, io.SeekStart); err != nil {
			return nil, err
		}
	} else {
		data := make([]byte, size)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, fmt.Errorf("chunk %q: read data (size=%d): %w", id, size, err)
		}
		chunk.Data = data
		if size%2 != 0 {
			if _, err := r.Seek(1, io.SeekCurrent); err != nil {
				return nil, err
			}
		}
	}
	return chunk, nil
}
