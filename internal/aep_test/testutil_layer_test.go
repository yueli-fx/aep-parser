package aep_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Layer / composition wrappers around an inner tdgp body. These return
// full RIFX bytes suitable for aep.FromReader.

// wrapAsLayer takes a fully-formed tdgp body (already including ADBE Group
// End) and produces a minimal one-layer AEP file containing exactly it.
func wrapAsLayer(tdgpBody []byte) []byte {
	rb := &rifxBuilder{}
	tdgpList := rb.listChunk("LIST", "tdgp", tdgpBody)
	ldtaData := make([]byte, 0x2C)
	binary.BigEndian.PutUint32(ldtaData[0x28:], 1)
	ldta := rb.chunk("ldta", ldtaData)
	var layerBody []byte
	layerBody = append(layerBody, ldta...)
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	compName := rb.chunk("Utf8", []byte("Test"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 300))
	var compBody []byte
	compBody = append(compBody, compName...)
	compBody = append(compBody, compIdta...)
	compBody = append(compBody, compCdta...)
	compBody = append(compBody, layrList...)
	compItem := rb.listChunk("LIST", "Item", compBody)
	foldList := rb.listChunk("LIST", "Fold", compItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

// ldtaOpts holds every ldta field buildLdta160 writes; used by per-field
// decoder tests that need precise control over each byte.
type ldtaOpts struct {
	LayerID, ParentID, SourceID  uint32
	Quality                      uint16
	StartTime, InPoint, OutPoint float64
	StretchDividend              int32
	StretchDivisor               uint32
	Flag25, Flag26, Flag27       byte
	Label                        byte
	BlendingMode                 byte
	TrackMatte                   byte
	PreserveTransparency         bool
	TypeByte                     byte
}

// buildLdta160 builds a 160-byte ldta that exercises every field
// parse_layer.go decodes. Tick rate is fixed at 8000 (matches the
// default cdta from buildCdta) so the time fields produce clean seconds.
func buildLdta160(opts ldtaOpts) []byte {
	d := make([]byte, 160)
	binary.BigEndian.PutUint32(d[0x00:], opts.LayerID)
	binary.BigEndian.PutUint16(d[0x04:], opts.Quality)
	binary.BigEndian.PutUint32(d[0x08:], uint32(opts.StretchDividend))
	binary.BigEndian.PutUint32(d[0x0C:], uint32(int32(opts.StartTime*8000)))
	binary.BigEndian.PutUint32(d[0x10:], 8000)
	binary.BigEndian.PutUint32(d[0x14:], uint32(int32(opts.InPoint*8000)))
	binary.BigEndian.PutUint32(d[0x18:], 8000)
	binary.BigEndian.PutUint32(d[0x1C:], uint32(int32(opts.OutPoint*8000)))
	binary.BigEndian.PutUint32(d[0x20:], 8000)
	d[0x25] = opts.Flag25
	d[0x26] = opts.Flag26
	d[0x27] = opts.Flag27
	binary.BigEndian.PutUint32(d[0x28:], opts.SourceID)
	d[0x3D] = opts.Label
	d[0x63] = opts.BlendingMode
	if opts.PreserveTransparency {
		d[0x67] = 1
	}
	d[0x6B] = opts.TrackMatte
	binary.BigEndian.PutUint32(d[0x6C:], opts.StretchDivisor)
	d[0x83] = opts.TypeByte
	binary.BigEndian.PutUint32(d[0x84:], opts.ParentID)
	return d
}

// wrapLayerInComp puts a single layer (the supplied ldta + property body
// + optional extra sibling chunks like cmta) into a comp and returns full
// RIFX bytes for parsing.
func wrapLayerInComp(rb *rifxBuilder, ldta, tdgpList []byte, extraSiblings ...[]byte) []byte {
	ldtaChunk := rb.chunk("ldta", ldta)
	var layerBody []byte
	layerBody = append(layerBody, ldtaChunk...)
	for _, s := range extraSiblings {
		layerBody = append(layerBody, s...)
	}
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	compName := rb.chunk("Utf8", []byte("T"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 100))
	var compBody []byte
	compBody = append(compBody, compName...)
	compBody = append(compBody, compIdta...)
	compBody = append(compBody, compCdta...)
	compBody = append(compBody, layrList...)
	compItem := rb.listChunk("LIST", "Item", compBody)
	foldList := rb.listChunk("LIST", "Fold", compItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

// buildMultiLayerComp wraps N ldta payloads as N Layr siblings inside one
// composition and returns full RIFX bytes. Each layer gets an empty tdgp
// (just an ADBE Group End sentinel) so parseProperties is a no-op.
func buildMultiLayerComp(rb *rifxBuilder, ldtas [][]byte) []byte {
	emptyTdgp := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))

	compName := rb.chunk("Utf8", []byte("T"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 100))
	var compBody []byte
	compBody = append(compBody, compName...)
	compBody = append(compBody, compIdta...)
	compBody = append(compBody, compCdta...)
	for _, ldta := range ldtas {
		ldtaChunk := rb.chunk("ldta", ldta)
		var layerBody []byte
		layerBody = append(layerBody, ldtaChunk...)
		layerBody = append(layerBody, emptyTdgp...)
		compBody = append(compBody, rb.listChunk("LIST", "Layr", layerBody)...)
	}
	compItem := rb.listChunk("LIST", "Item", compBody)
	foldList := rb.listChunk("LIST", "Fold", compItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

// buildMultiCompProject builds a project containing N comps + 1 footage
// item. The first comp receives the supplied layers (each layer is just
// an ldta + empty tdgp), giving callers a layer set whose SourceIDs can
// reference the other comp IDs (2, 3, ...) or the footage ID (lastID+1).
//
// Comp IDs are 1, 2, 3, ... and the footage gets the next ID after the
// last comp.
func buildMultiCompProject(rb *rifxBuilder, compCount int, primaryCompLayers [][]byte) []byte {
	emptyTdgp := rb.listChunk("LIST", "tdgp", rb.chunk("tdmn", []byte("ADBE Group End")))

	var foldData []byte
	for i := 1; i <= compCount; i++ {
		var compBody []byte
		compBody = append(compBody, rb.chunk("Utf8", []byte(fmt.Sprintf("Comp%d", i)))...)
		compBody = append(compBody, rb.chunk("idta", buildIdta(0x04, uint32(i)))...)
		compBody = append(compBody, rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 100))...)
		if i == 1 {
			for _, ldta := range primaryCompLayers {
				var layerBody []byte
				layerBody = append(layerBody, rb.chunk("ldta", ldta)...)
				layerBody = append(layerBody, emptyTdgp...)
				compBody = append(compBody, rb.listChunk("LIST", "Layr", layerBody)...)
			}
		}
		foldData = append(foldData, rb.listChunk("LIST", "Item", compBody)...)
	}

	footageID := uint32(compCount + 1)
	footageIdta := rb.chunk("idta", buildIdta(0x07, footageID))
	sspcData := make([]byte, 4)
	binary.BigEndian.PutUint16(sspcData[0:], 1280)
	binary.BigEndian.PutUint16(sspcData[2:], 720)
	var pinData []byte
	pinData = append(pinData, rb.chunk("sspc", sspcData)...)
	optiData := make([]byte, 0x1A)
	copy(optiData[:4], "Soli")
	optiData = append(optiData, []byte("bg.mp4\x00")...)
	pinData = append(pinData, rb.chunk("opti", optiData)...)
	pinList := rb.listChunk("LIST", "Pin ", pinData)
	var footageBody []byte
	footageBody = append(footageBody, footageIdta...)
	footageBody = append(footageBody, pinList...)
	foldData = append(foldData, rb.listChunk("LIST", "Item", footageBody)...)

	foldList := rb.listChunk("LIST", "Fold", foldData)
	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

// buildLdtaFullLengthAEP wraps a layer with a 160-byte ldta (rather than
// the buildExtendedAEP minimal 0x2C) so flag-bit and byte-field setters
// have valid offsets to write to. Uses the standard rifxBuilder helpers
// and seeds ldta @0x27 bit0 (Visible) on so the parser produces the
// usual default (visible-by-default) layer.
func buildLdtaFullLengthAEP() []byte {
	rb := &rifxBuilder{}
	ldtaData := make([]byte, 160)
	binary.BigEndian.PutUint32(ldtaData[0x28:], 1) // SourceID = 1
	ldtaData[0x27] = 0x01                          // Visible bit
	ldta := rb.chunk("ldta", ldtaData)

	tdgpBody := rb.chunk("tdmn", []byte("ADBE Group End"))
	tdgpList := rb.listChunk("LIST", "tdgp", tdgpBody)

	var layerBody []byte
	layerBody = append(layerBody, ldta...)
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	compName := rb.chunk("Utf8", []byte("Comp"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 300))
	var compBody []byte
	compBody = append(compBody, compName...)
	compBody = append(compBody, compIdta...)
	compBody = append(compBody, compCdta...)
	compBody = append(compBody, layrList...)
	compItem := rb.listChunk("LIST", "Item", compBody)
	foldList := rb.listChunk("LIST", "Fold", compItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

