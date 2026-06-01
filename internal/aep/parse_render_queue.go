package aep

import (
	"encoding/binary"
	"encoding/json"

	"github.com/example/aep-parser/internal/rifx"
)

// parse_render_queue.go — parse the render queue (LIST:LRdr) into the scene
// model. Read-only (P3 §3A slice-1). Structure (verified by tmp_debug probe):
//
//	LRdr
//	 ├ LIST:list → lhd3 + ldat(RenderSettingsItem × N, 2246B)   ← per-item settings
//	 └ LIST:LItm → per item: [RCom] + LIST:list(om settings) + LIST:'LOm '
//
// LItm items are matched by walking children: a 'LOm ' LIST closes the item
// started by the preceding 'list' (and optional preceding RCom comment).

func parseRenderQueue(root *rifx.Chunk, proj *Project) {
	lrdr := findFirstListDeep(root, rifx.IDLRdr)
	if lrdr == nil {
		return
	}
	rq := &RenderQueue{}
	proj.RenderQueue = rq

	settingsBlocks := renderSettingsBlocks(lrdr)
	litm := lrdr.FindFirstList(rifx.IDLItm)
	if litm == nil {
		return
	}

	idx := 0
	var pendingComment string
	haveList := false
	for _, ch := range litm.Children {
		switch {
		case ch.ID == rifx.IDRCom:
			pendingComment = decodeRComComment(ch.Data)
		case ch.IsList() && ch.FormType == rifx.IDkfl:
			haveList = true
		case ch.IsList() && ch.FormType == rifx.IDLOm:
			if !haveList {
				continue
			}
			item := buildRenderQueueItem(settingsBlocks, idx, pendingComment, ch, proj)
			rq.Items = append(rq.Items, item)
			idx++
			pendingComment = ""
			haveList = false
		}
	}
}

// renderSettingsBlocks returns the raw per-item settings ldat split into
// 2246-byte blocks (one per render queue item). Nil when absent/empty.
func renderSettingsBlocks(lrdr *rifx.Chunk) [][]byte {
	list := lrdr.FindFirstList(rifx.IDkfl)
	if list == nil {
		return nil
	}
	ldat := list.FindFirst(rifx.IDLdat)
	if ldat == nil {
		return nil
	}
	n := len(ldat.Data) / renderSettingsItemSize
	blocks := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		blocks = append(blocks, ldat.Data[i*renderSettingsItemSize:(i+1)*renderSettingsItemSize])
	}
	return blocks
}

func buildRenderQueueItem(blocks [][]byte, idx int, comment string, lom *rifx.Chunk, proj *Project) *RenderQueueItem {
	item := &RenderQueueItem{Comment: comment}
	if idx < len(blocks) {
		if rs, ok := decodeRenderSettings(blocks[idx]); ok {
			item.Status = rs.status
			item.Name = rs.templateName
			item.Comp = proj.CompositionByID(rs.compID)
			item.TimeSpanStart, item.TimeSpanDuration = resolveTimeSpan(rs, item.Comp)
		}
	}
	item.OutputModules = parseOutputModules(lom)
	return item
}

// resolveTimeSpan turns the time_span_source + dividends into (start, duration)
// seconds, mirroring py-aep RenderQueueItem._resolved_time_span.
func resolveTimeSpan(rs renderSettings, comp *Composition) (start, dur float64) {
	switch rs.timeSpanSource {
	case timeSpanLengthOfComp:
		if comp != nil {
			return 0, comp.Duration
		}
		return 0, 0
	case timeSpanWorkAreaOnly:
		if comp != nil {
			return comp.WorkAreaStart, comp.WorkAreaEnd - comp.WorkAreaStart
		}
		return 0, 0
	default: // CUSTOM (2 or 0xFFFF)
		return ratio(rs.tsStartDividend, rs.tsStartDivisor), ratio(rs.tsDurationDivdend, rs.tsDurationDivisor)
	}
}

func ratio(dividend, divisor uint32) float64 {
	if divisor == 0 {
		return 0
	}
	return float64(dividend) / float64(divisor)
}

// parseOutputModules splits a 'LOm ' LIST into per-module groups (each starts
// at a Roou chunk) and extracts the read-only fields for each.
func parseOutputModules(lom *rifx.Chunk) []*OutputModule {
	var oms []*OutputModule
	var group []*rifx.Chunk
	flush := func() {
		if len(group) > 0 {
			oms = append(oms, buildOutputModule(group))
			group = nil
		}
	}
	for _, ch := range lom.Children {
		if ch.ID == rifx.IDRoou {
			flush()
		}
		group = append(group, ch)
	}
	flush()
	return oms
}

func buildOutputModule(group []*rifx.Chunk) *OutputModule {
	om := &OutputModule{}
	als2Seen := false
	var postAls2 []string
	for _, ch := range group {
		if ch.IsList() && ch.FormType == rifx.IDAls2 {
			als2Seen = true
			if alas := ch.FindFirst(rifx.IDAlas); alas != nil {
				om.FullPath = decodeAlasFullPath(alas.Data)
			}
			continue
		}
		if als2Seen && ch.ID == rifx.IDUtf8 {
			postAls2 = append(postAls2, ch.Text())
		}
	}
	if len(postAls2) > 0 {
		om.Name = postAls2[0]
	}
	if len(postAls2) > 1 {
		om.FileTemplate = postAls2[1]
	}
	return om
}

// decodeRComComment extracts the comment string from an RCom leaf. RCom is a
// non-LIST container whose Data holds an embedded Utf8 chunk
// ("Utf8" + u32 len + payload).
func decodeRComComment(data []byte) string {
	if len(data) < 8 || string(data[0:4]) != "Utf8" {
		return ""
	}
	n := int(binary.BigEndian.Uint32(data[4:8]))
	if 8+n > len(data) {
		n = len(data) - 8
	}
	return trimNUL(data[8 : 8+n])
}

func decodeAlasFullPath(data []byte) string {
	var alias struct {
		FullPath string `json:"fullpath"`
	}
	if err := json.Unmarshal(data, &alias); err == nil {
		return alias.FullPath
	}
	return ""
}

// findFirstListDeep returns the first LIST (any depth, pre-order) with the
// given formType, or nil.
func findFirstListDeep(c *rifx.Chunk, ft rifx.ChunkID) *rifx.Chunk {
	for _, ch := range c.Children {
		if ch.IsList() && ch.FormType == ft {
			return ch
		}
		if found := findFirstListDeep(ch, ft); found != nil {
			return found
		}
	}
	return nil
}
