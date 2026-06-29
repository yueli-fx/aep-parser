package serializer

import (
	"encoding/binary"
	"encoding/json"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// parse_render_queue.go — parse the render queue (LIST:LRdr) into the scene
// model. Read-only (P3 §3A slice-1). Structure (verified by a chunk-tree probe):
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
	scene.SetRenderQueueBack(rq, &renderQueueBackrefs{lrdr: lrdr})
	proj.RenderQueue = rq

	settingsBlocks := renderSettingsBlocks(lrdr)
	litm := lrdr.FindFirstList(rifx.IDLItm)
	if litm == nil {
		return
	}

	idx := 0
	var pendingComment string
	var pendingRcom, pendingList *rifx.Chunk
	for _, ch := range litm.Children {
		switch {
		case ch.ID == rifx.IDRCom:
			pendingComment = decodeRComComment(ch.Data)
			pendingRcom = ch
		case ch.IsList() && ch.FormType == rifx.IDkfl:
			pendingList = ch
		case ch.IsList() && ch.FormType == rifx.IDLOm:
			if pendingList == nil {
				continue
			}
			item := buildRenderQueueItem(settingsBlocks, idx, pendingComment, litm, pendingList, ch, pendingRcom, proj)
			rq.Items = append(rq.Items, item)
			idx++
			pendingComment = ""
			pendingRcom = nil
			pendingList = nil
		}
	}
}

// outputModuleSettingsBlocks splits an LItm-item's LIST:list ldat into 128-byte
// blocks, one per output module. Nil when absent/empty.
func outputModuleSettingsBlocks(itemList *rifx.Chunk) [][]byte {
	ldat := itemList.FindFirst(rifx.IDLdat)
	if ldat == nil {
		return nil
	}
	n := len(ldat.Data) / codec.OutputModuleSettingsItemSize
	blocks := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		blocks = append(blocks, ldat.Data[i*codec.OutputModuleSettingsItemSize:(i+1)*codec.OutputModuleSettingsItemSize])
	}
	return blocks
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
	n := len(ldat.Data) / codec.RenderSettingsItemSize
	blocks := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		blocks = append(blocks, ldat.Data[i*codec.RenderSettingsItemSize:(i+1)*codec.RenderSettingsItemSize])
	}
	return blocks
}

func buildRenderQueueItem(blocks [][]byte, idx int, comment string, litm, itemList, lom, rcom *rifx.Chunk, proj *Project) *RenderQueueItem {
	item := &RenderQueueItem{Comment: comment}
	var settingsAlias []byte
	if idx < len(blocks) {
		settingsAlias = blocks[idx]
		scene.SetRenderQueueItemSettings(item, append([]byte(nil), blocks[idx]...))
		if rs, ok := codec.DecodeRenderSettings(blocks[idx]); ok {
			item.Status = rs.Status
			item.Name = rs.TemplateName
			item.Comp = proj.CompositionByID(rs.CompID)
			item.TimeSpanStart, item.TimeSpanDuration = resolveTimeSpan(rs, item.Comp)
			item.LogType = rs.LogType
			item.QueueItemNotify = rs.QueueItemNotify
			item.ElapsedSeconds = rs.ElapsedSeconds
			item.RenderSettings = RenderSettings{
				Quality:           codec.CurrentSettingsInt(rs.Quality),
				ColorDepth:        codec.CurrentSettingsInt(rs.ColorDepth),
				Effects:           codec.CurrentSettingsInt(rs.Effects),
				FieldRender:       int(rs.FieldRender),
				Pulldown:          int(rs.Pulldown),
				FrameBlending:     codec.CurrentSettingsInt(rs.FrameBlending),
				MotionBlur:        codec.CurrentSettingsInt(rs.MotionBlur),
				ProxyUse:          codec.CurrentSettingsInt(rs.ProxyUse),
				SoloSwitches:      codec.CurrentSettingsInt(rs.SoloSwitches),
				GuideLayers:       codec.CurrentSettingsInt(rs.GuideLayers),
				DiskCache:         codec.CurrentSettingsInt(rs.DiskCache),
				FrameRate:         int(rs.UseThisFrameRate),
				Resolution:        [2]int{int(rs.ResolutionX), int(rs.ResolutionY)},
				SkipExistingFiles: rs.SkipExistingFiles,
			}
		}
	}
	scene.SetRenderQueueItemBack(item, &renderQueueItemBackrefs{
		litm:          litm,
		itemListChunk: itemList,
		rcomChunk:     rcom,
		settingsSlice: settingsAlias,
	})
	item.OutputModules = parseOutputModules(lom, outputModuleSettingsBlocks(itemList))
	return item
}

// resolveTimeSpan turns the time_span_source + dividends into (start, duration)
// seconds, mirroring the reference parser's RenderQueueItem._resolved_time_span.
func resolveTimeSpan(rs codec.RenderSettingsBlock, comp *Composition) (start, dur float64) {
	switch rs.TimeSpanSource {
	case codec.TimeSpanLengthOfComp:
		if comp != nil {
			return 0, comp.Duration
		}
		return 0, 0
	case codec.TimeSpanWorkAreaOnly:
		if comp != nil {
			return comp.WorkAreaStart, comp.WorkAreaEnd - comp.WorkAreaStart
		}
		return 0, 0
	default: // CUSTOM (2 or 0xFFFF)
		return ratio(rs.TsStartDividend, rs.TsStartDivisor), ratio(rs.TsDurationDivdend, rs.TsDurationDivisor)
	}
}

func ratio(dividend, divisor uint32) float64 {
	if divisor == 0 {
		return 0
	}
	return float64(dividend) / float64(divisor)
}

// parseOutputModules splits a 'LOm ' LIST into per-module groups (each starts
// at a Roou chunk) and extracts the read-only fields for each. omBlocks holds
// the per-module 128B settings, paired by module index.
func parseOutputModules(lom *rifx.Chunk, omBlocks [][]byte) []*OutputModule {
	var oms []*OutputModule
	var group []*rifx.Chunk
	flush := func() {
		if len(group) > 0 {
			var block []byte
			if len(oms) < len(omBlocks) {
				block = omBlocks[len(oms)]
			}
			oms = append(oms, buildOutputModule(group, block))
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

func buildOutputModule(group []*rifx.Chunk, omBlock []byte) *OutputModule {
	omb := &outputModuleBackrefs{settingsSlice: omBlock}
	om := &OutputModule{}
	scene.SetOutputModuleSettingsBlock(om, append([]byte(nil), omBlock...))
	scene.SetOutputModuleBack(om, omb)
	als2Seen := false
	var postAls2 []string
	for _, ch := range group {
		if ch.ID == rifx.IDRoou {
			scene.SetOutputModuleRoouData(om, append([]byte(nil), ch.Data...))
			omb.roouSlice = ch.Data
			applyRoou(om, ch.Data)
			continue
		}
		if ch.ID == rifx.IDRopt {
			om.FormatOptions = codec.DecodeRoptFormatOptions(ch.Data)
			continue
		}
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
	if s, ok := codec.DecodeOMSettings(omBlock); ok {
		om.Settings.Channels = s.Channels
		om.Settings.ResizeQuality = s.ResizeQuality
		om.Settings.Resize = s.Resize
		om.Settings.LockAspectRatio = s.LockAspectRatio
		om.Settings.Crop = s.Crop
		om.Settings.CropTop = s.CropTop
		om.Settings.CropLeft = s.CropLeft
		om.Settings.CropBottom = s.CropBottom
		om.Settings.CropRight = s.CropRight
		om.Settings.OutputAudio = s.OutputAudio
		om.Settings.IncludeProjectLink = s.IncludeProjectLink
		om.Settings.PostRenderAction = s.PostRenderAction
		om.Settings.ConvertToLinear = s.ConvertToLinear
		om.Settings.UseCompFrameNumber = s.UseCompFrameNumber
		om.Settings.UseRegionOfInterest = s.UseRegionOfInterest
		om.Settings.IncludeSourceXMP = s.IncludeSourceXMP
		om.Settings.PreserveRGB = s.PreserveRGB
	}
	return om
}

func applyRoou(om *OutputModule, data []byte) {
	r, ok := codec.DecodeRoou(data)
	if !ok {
		return
	}
	om.Settings.VideoCodec = r.VideoCodec
	om.Settings.FormatID = r.FormatID
	om.Settings.StartingNumber = r.StartingNumber
	om.Settings.Width = r.Width
	om.Settings.Height = r.Height
	om.Settings.Depth = r.Depth
	om.Settings.VideoOutput = r.Width > 0 || r.Height > 0
	om.Settings.AudioSampleRate = r.AudioSampleRate
	om.Settings.AudioBitDepth = r.AudioBitDepth
	om.Settings.AudioChannels = r.AudioChannels
	om.Settings.AudioEnabled = r.AudioEnabled
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
