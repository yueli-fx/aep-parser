package aep

import (
	"encoding/binary"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// parseMarkers decodes the "ADBE Marker" mrst wrapper into one Marker per
// keyframe. Structure (observed):
//
//	mrst LIST
//	├── tdbs LIST { tdb4, kfl LIST { lhd3 (count, bpk=16), ldat (count*16) } }
//	│       Per-block: 0x00-0x03 = time (1/8000 ticks); 0x04-0x0F = opaque metadata
//	└── mrky LIST { Nmrd LIST per marker { NmHd + 5×Utf8: comment/chapter/URL/frame/cue } }
//
// Times come from the tdbs ldat (lhd3@0x10 = bpk=16 fixed). Text fields come
// from each Nmrd's 5 Utf8 children, in declaration order. Duration + label
// color come from the NmHd chunk inside each Nmrd — see decodeNmHd.
func parseMarkers(mrst *rifx.Chunk, ctx *parseCtx) []*Marker {
	var tdbs, mrky *rifx.Chunk
	for _, ch := range mrst.Children {
		if !ch.IsList() {
			continue
		}
		switch ch.FormType {
		case rifx.IDTdbs:
			tdbs = ch
		case rifx.IDMrky:
			mrky = ch
		}
	}
	if tdbs == nil {
		return nil
	}
	var kfl *rifx.Chunk
	for _, ch := range tdbs.Children {
		if ch.IsList() && ch.FormType == rifx.IDkfl {
			kfl = ch
			break
		}
	}
	if kfl == nil {
		return nil
	}
	lhd3 := kfl.FindFirst(rifx.IDLhd3)
	ldat := kfl.FindFirst(rifx.IDLdat)
	if lhd3 == nil || ldat == nil || len(lhd3.Data) < 0x14 {
		return nil
	}
	count := int(binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	if count <= 0 || bpk <= 0 || count*bpk > len(ldat.Data) {
		return nil
	}

	// Gather Nmrd siblings from mrky in order — one per marker.
	var nmrds []*rifx.Chunk
	if mrky != nil {
		for _, ch := range mrky.Children {
			if ch.IsList() && ch.FormType == rifx.IDNmrd {
				nmrds = append(nmrds, ch)
			}
		}
	}

	// Shared container for structural ops (Remove/Add). owner is wired later
	// by bindMarkerListOwner once the slice has been assigned to its public
	// Composition.Markers / Layer.Markers field.
	ml := &markerList{lhd3: lhd3, ldat: ldat, mrky: mrky}

	markers := make([]*Marker, 0, count)
	for i := 0; i < count; i++ {
		off := i * bpk
		mb := &markerBackrefs{ldat: ldat, ldatOffset: off, tickRate: ctx.tickRate}
		m := &Marker{
			Time: float64(binary.BigEndian.Uint32(ldat.Data[off:off+4])) / ctx.tickRate,
		}
		scene.SetMarkerLdatOffset(m, off)
		scene.SetMarkerRates(m, ctx.tickRate, ctx.compFps)
		scene.SetMarkerBack(m, mb)
		scene.SetMarkerSetList(m, ml)
		if i < len(nmrds) {
			mb.nmrd = nmrds[i]
			fillMarkerText(m, nmrds[i])
			if nmHd := nmrds[i].FindFirst(rifx.IDNmhd); nmHd != nil {
				mb.nmHd = nmHd
				decodeNmHd(m, nmHd.Data)
			}
		}
		markers = append(markers, m)
	}
	return markers
}

// bindMarkerListOwner records, on the shared markerList behind a parsed marker
// set, a pointer to the public []*Marker field the set lives in (a comp's or a
// layer's Markers). Structural ops (Remove/Add) use it to keep that slice in
// sync with the chunk tree. No-op for an empty set. Call once, right after the
// slice has been assigned to its owning field.
func bindMarkerListOwner(field *[]*Marker) {
	for _, m := range *field {
		if ml := markerSet(m); ml != nil {
			ml.owner = field
			return
		}
	}
}

// decodeNmHd populates the marker's Duration + Label from a NmHd chunk
// payload.
//
// Sources:
//   - "py-aep" = py-aep/binary/misc_chunks.py NmhdChunk
//   - "fixture" = hex-dumped from test_data/re_tickrate.aep
//
// Layout (py-aep verified; fixture is consistent for zero/label only —
// see ⚠ note below):
//
//	0x00..0x02 : reserved (3 bytes)                                          [py-aep + fixture]
//	0x03       : marker_flags (uint8) — bit0=navigation, bit1=protected     [py-aep] — decode deferred this loop
//	0x04..0x07 : reserved (4 bytes)                                          [py-aep + fixture]
//	0x08..0x0B : frame_duration (uint32 BE) — 600ths of a second per py-aep  [py-aep stated; fixture observes 0 here in 2/2 point markers]
//	0x0C..0x0F : reserved per py-aep                                         [⚠ py-aep + fixture DISAGREE — see note]
//	0x10       : label color index (uint8, 0..16; AE's user-customizable    [py-aep + fixture]
//	             label palette — index is stable, RGB mapping is not)
//	0x11+      : optional trailing padding (3 bytes in CC2020+ fixtures)     [local]
//
// ⚠ Offset ambiguity (do NOT speculate, do NOT change decode):
// Both NmHd chunks in re_tickrate.aep (CC2020+ markers) have:
//
//	bytes 0x08..0x0B = 00 00 00 00  (duration per py-aep — zero)
//	bytes 0x0C..0x0F = 00 00 02 58  (= 600; py-aep says "reserved")
//
// 0x258 = 600 = exactly one second at the py-aep documented 600-base.
// This is highly suspicious but unverified — could mean (1) py-aep's
// duration offset shifted in newer AE versions to 0x0C, (2) the field
// is a related-but-distinct semantic (e.g. an internal "1 unit" sentinel
// for chapter markers), or (3) coincidence. Until we have an AE
// project with a known user-set duration to verify against, we follow
// py-aep's documented offset (0x08). Real-world point markers will
// correctly surface as Duration = 0.0s under this decode.
//
// py-aep declares NmHd as 17 bytes; real-world fixtures pad to 20. We
// bounds-check both reads.
func decodeNmHd(m *Marker, d []byte) {
	if len(d) >= 0x0C {
		// 600ths-of-a-second per py-aep; fixture's 0x258 = 600 = 1.0s
		// is plausible "1-second marker". Surface as seconds for
		// consistency with Marker.Time.
		const nmHdDurationBase = 600.0
		m.Duration = float64(binary.BigEndian.Uint32(d[0x08:0x0C])) / nmHdDurationBase
	}
	if len(d) >= 0x11 {
		m.Label = d[0x10]
	}
}

// fillMarkerText pulls the first 5 Utf8 children (in declaration order) into
// the Marker's text fields. Missing Utf8 chunks leave the field empty.
func fillMarkerText(m *Marker, nmrd *rifx.Chunk) {
	var utf8s []string
	for _, ch := range nmrd.Children {
		if ch.ID == rifx.IDUtf8 {
			utf8s = append(utf8s, ch.Text())
		}
	}
	if len(utf8s) >= 1 {
		m.Comment = utf8s[0]
	}
	if len(utf8s) >= 2 {
		m.Chapter = utf8s[1]
	}
	if len(utf8s) >= 3 {
		m.URL = utf8s[2]
	}
	if len(utf8s) >= 4 {
		m.FrameTarget = utf8s[3]
	}
	if len(utf8s) >= 5 {
		m.CuePointName = utf8s[4]
	}
}
