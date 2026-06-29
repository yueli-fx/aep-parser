package aep_test

import (
	"bytes"
	"encoding/binary"
)

// Synthetic RIFX builder shared by all tests. Per-test specialized builders
// (build*Keyframed, build*MaskAtom*, leaf*, buildKF*, wrapAsLayer, etc.)
// live alongside their respective tests in aep_test.go — they're too
// tightly coupled to specific keyframe / mask layout encoding to be
// useful elsewhere.

type rifxBuilder struct {
	buf bytes.Buffer
}

// chunk emits a single leaf chunk: 4-byte ID, BE uint32 size, payload,
// optional pad byte to make the total slot size even.
func (b *rifxBuilder) chunk(id string, data []byte) []byte {
	var out bytes.Buffer
	out.WriteString(id)
	_ = binary.Write(&out, binary.BigEndian, uint32(len(data)))
	out.Write(data)
	if len(data)%2 != 0 {
		out.WriteByte(0) // padding
	}
	return out.Bytes()
}

// listChunk emits a LIST/RIFX chunk: 4-byte ID, BE uint32 size, 4-byte
// formType, then the pre-formed children bytes.
func (b *rifxBuilder) listChunk(id, formType string, children []byte) []byte {
	var out bytes.Buffer
	out.WriteString(id)
	_ = binary.Write(&out, binary.BigEndian, uint32(len(formType)+len(children)))
	out.WriteString(formType)
	out.Write(children)
	return out.Bytes()
}

// buildIdta creates an idta block matching boltframe's IDTA layout:
// Type uint16 @ 0x00, ID uint32 @ 0x10. typeRaw: 1=folder, 4=comp, 7=footage.
func buildIdta(typeRaw uint16, itemID uint32) []byte {
	d := make([]byte, 20)
	binary.BigEndian.PutUint16(d[0:], typeRaw)
	binary.BigEndian.PutUint32(d[16:], itemID)
	return d
}

// buildCdta creates a composition data block with the offsets verified
// against a real modern .aep file:
//
//	Width  @ 0x8C (uint16), Height @ 0x8E (uint16)
//	FramerateWhole @ 0x9C (uint16), FramerateFractional @ 0x9E (uint16)
//	DurationFrames @ 0xB0 (uint32)
//
// The default cdta_0x08 = 8000 with cdta_0xA8 = 1 → TickRate = 8000.
// This matches the synthetic tests which encode keyframe times at
// 1/8000-second ticks (via uint32(time*8000) in buildKF*).
func buildCdta(width, height, fpsWhole uint16, fpsFrac uint16, durFrames uint32) []byte {
	// 204 bytes matches the real-world cdta size (CC2020+). The shutter /
	// motion blur fields live in the 0xAE..0xCC tail, beyond the legacy
	// 0xB4-byte buffer we used before.
	d := make([]byte, 204)
	binary.BigEndian.PutUint16(d[0x8C:], width)
	binary.BigEndian.PutUint16(d[0x8E:], height)
	binary.BigEndian.PutUint16(d[0x9C:], fpsWhole)
	binary.BigEndian.PutUint16(d[0x9E:], fpsFrac)
	binary.BigEndian.PutUint32(d[0xB0:], durFrames)
	binary.BigEndian.PutUint32(d[0x08:], 8000)
	binary.BigEndian.PutUint32(d[0xA8:], 1)
	return d
}

// cdtaOpts holds the cdta-tail fields decoded by parseComposition's CDTA
// expansion. Zero values map to the AE defaults the parser would otherwise
// see for an unset/missing field, EXCEPT for the work-area-end sentinel:
// callers should pass WorkAreaEndDividend = 0xFFFFFFFF to exercise the
// "extend to comp Duration" substitution path.
type cdtaOpts struct {
	BGColor                   [3]uint8
	WorkAreaStartDividend     uint32
	WorkAreaStartDivisor      uint32
	WorkAreaEndDividend       uint32
	WorkAreaEndDivisor        uint32
	ShutterAngle              uint16
	ShutterPhase              int32
	MotionBlurAdaptiveLimit   int32
	MotionBlurSamplesPerFrame int32
}

// buildCdtaWithOpts is buildCdta plus an opts struct that writes every
// CDTA tail field the parser decodes. Used by TestCompositionCDTAFields.
func buildCdtaWithOpts(width, height, fpsWhole, fpsFrac uint16, durFrames uint32, opts cdtaOpts) []byte {
	d := buildCdta(width, height, fpsWhole, fpsFrac, durFrames)
	d[0x34] = opts.BGColor[0]
	d[0x35] = opts.BGColor[1]
	d[0x36] = opts.BGColor[2]
	binary.BigEndian.PutUint32(d[0x1C:], opts.WorkAreaStartDividend)
	binary.BigEndian.PutUint32(d[0x20:], opts.WorkAreaStartDivisor)
	binary.BigEndian.PutUint32(d[0x24:], opts.WorkAreaEndDividend)
	binary.BigEndian.PutUint32(d[0x28:], opts.WorkAreaEndDivisor)
	binary.BigEndian.PutUint16(d[0xAE:], opts.ShutterAngle)
	binary.BigEndian.PutUint32(d[0xB4:], uint32(opts.ShutterPhase))
	binary.BigEndian.PutUint32(d[0xC4:], uint32(opts.MotionBlurAdaptiveLimit))
	binary.BigEndian.PutUint32(d[0xC8:], uint32(opts.MotionBlurSamplesPerFrame))
	return d
}

// buildCdtaWithRate is buildCdta with explicit cdta_0x08 (tick rate) and
// cdta_0xA8 (legacy scale factor). Tests that exercise non-default tick
// rates use this.
func buildCdtaWithRate(width, height, fpsWhole, fpsFrac uint16, durFrames, tickRate, legacyScale uint32) []byte {
	d := buildCdta(width, height, fpsWhole, fpsFrac, durFrames)
	binary.BigEndian.PutUint32(d[0x08:], tickRate)
	binary.BigEndian.PutUint32(d[0xA8:], legacyScale)
	return d
}

// buildMinimalAEP creates a minimal valid RIFX AEP file with one composition,
// one footage, and one folder. Used by the smoke / regression tests.
func buildMinimalAEP() []byte {
	rb := &rifxBuilder{}

	// Composition (idta type 0x04): 1920x1080, 30 fps, 10s (300 frames).
	compName := rb.chunk("Utf8", []byte("Test Comp"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 300))
	var compData []byte
	compData = append(compData, compName...)
	compData = append(compData, compIdta...)
	compData = append(compData, compCdta...)
	compItem := rb.listChunk("LIST", "Item", compData)

	// Footage (idta type 0x07) wrapped in a Pin sublist with sspc + opti.
	footageIdta := rb.chunk("idta", buildIdta(0x07, 2))
	sspcData := make([]byte, 4)
	binary.BigEndian.PutUint16(sspcData[0:], 1280)
	binary.BigEndian.PutUint16(sspcData[2:], 720)
	sspcChunk := rb.chunk("sspc", sspcData)
	// opti: 4-byte kind "Soli" then 22 bytes padding then NUL-terminated name.
	optiData := make([]byte, 0x1A)
	copy(optiData[:4], "Soli")
	optiData = append(optiData, []byte("bg.mp4\x00")...)
	optiChunk := rb.chunk("opti", optiData)
	var pinData []byte
	pinData = append(pinData, sspcChunk...)
	pinData = append(pinData, optiChunk...)
	pinList := rb.listChunk("LIST", "Pin ", pinData)

	var footageData []byte
	footageData = append(footageData, footageIdta...)
	footageData = append(footageData, pinList...)
	footageItem := rb.listChunk("LIST", "Item", footageData)

	// Folder (idta type 0x01).
	folderName := rb.chunk("Utf8", []byte("Solids"))
	folderIdta := rb.chunk("idta", buildIdta(0x01, 3))
	var folderData []byte
	folderData = append(folderData, folderName...)
	folderData = append(folderData, folderIdta...)
	folderItem := rb.listChunk("LIST", "Item", folderData)

	// Wrap the items in a Fold LIST as real .aep files do.
	var foldData []byte
	foldData = append(foldData, compItem...)
	foldData = append(foldData, footageItem...)
	foldData = append(foldData, folderItem...)
	foldList := rb.listChunk("LIST", "Fold", foldData)

	// Root Egg! content
	var eggData []byte
	eggData = append(eggData, foldList...)

	// RIFX root
	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(eggData)))
	root.WriteString("Egg!")
	root.Write(eggData)

	return root.Bytes()
}
