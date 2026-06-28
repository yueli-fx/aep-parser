package aeoracle

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const (
	FrameStatusOK              = "ok"
	FrameStatusDifferent       = "different"
	FrameStatusMissingExpected = "missing_expected"
	FrameStatusMissingActual   = "missing_actual"
)

type FrameSetCompareReport struct {
	SchemaVersion    int                  `json:"schema_version"`
	ExpectedMetadata string               `json:"expected_metadata"`
	ActualMetadata   string               `json:"actual_metadata"`
	ExpectedAEPPath  string               `json:"expected_aep_path"`
	ActualAEPPath    string               `json:"actual_aep_path"`
	CompName         string               `json:"comp_name"`
	Summary          FrameSetSummary      `json:"summary"`
	Frames           []FrameCompareRecord `json:"frames"`
}

type FrameSetSummary struct {
	TotalFrames           int `json:"total_frames"`
	OKFrames              int `json:"ok_frames"`
	DifferentFrames       int `json:"different_frames"`
	MissingExpectedFrames int `json:"missing_expected_frames"`
	MissingActualFrames   int `json:"missing_actual_frames"`
}

type FrameCompareRecord struct {
	Tag          string         `json:"tag"`
	Frame        int            `json:"frame"`
	Seconds      float64        `json:"seconds"`
	Reason       string         `json:"reason"`
	ExpectedPath string         `json:"expected_path,omitempty"`
	ActualPath   string         `json:"actual_path,omitempty"`
	Status       string         `json:"status"`
	Compare      *CompareReport `json:"compare,omitempty"`
	Error        string         `json:"error,omitempty"`
}

func ReadRenderMetadata(path string) (RenderMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RenderMetadata{}, err
	}
	var meta RenderMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return RenderMetadata{}, err
	}
	if meta.SchemaVersion != SchemaVersion {
		return RenderMetadata{}, fmt.Errorf("aeoracle: unsupported metadata schema_version %d", meta.SchemaVersion)
	}
	return meta, nil
}

func CompareFrameSets(expectedMetaPath, actualMetaPath string, opts CompareOptions) (FrameSetCompareReport, error) {
	expected, err := ReadRenderMetadata(expectedMetaPath)
	if err != nil {
		return FrameSetCompareReport{}, fmt.Errorf("aeoracle: read expected metadata: %w", err)
	}
	actual, err := ReadRenderMetadata(actualMetaPath)
	if err != nil {
		return FrameSetCompareReport{}, fmt.Errorf("aeoracle: read actual metadata: %w", err)
	}
	if expected.Status != "ok" {
		return FrameSetCompareReport{}, fmt.Errorf("aeoracle: expected metadata status %q", expected.Status)
	}
	if actual.Status != "ok" {
		return FrameSetCompareReport{}, fmt.Errorf("aeoracle: actual metadata status %q", actual.Status)
	}

	expectedFrames, err := indexRenderedFrames("expected", expected.Frames)
	if err != nil {
		return FrameSetCompareReport{}, err
	}
	actualFrames, err := indexRenderedFrames("actual", actual.Frames)
	if err != nil {
		return FrameSetCompareReport{}, err
	}

	report := FrameSetCompareReport{
		SchemaVersion:    SchemaVersion,
		ExpectedMetadata: expectedMetaPath,
		ActualMetadata:   actualMetaPath,
		ExpectedAEPPath:  expected.AEPPath,
		ActualAEPPath:    actual.AEPPath,
		CompName:         expected.CompName,
	}
	seenActual := map[string]bool{}
	for _, expectedFrame := range expected.Frames {
		actualFrame, ok := actualFrames[expectedFrame.Tag]
		record := FrameCompareRecord{
			Tag:          expectedFrame.Tag,
			Frame:        expectedFrame.Frame,
			Seconds:      expectedFrame.Seconds,
			Reason:       expectedFrame.Reason,
			ExpectedPath: expectedFrame.OutputPath,
		}
		if !ok {
			record.Status = FrameStatusMissingActual
			report.Frames = append(report.Frames, record)
			continue
		}
		seenActual[expectedFrame.Tag] = true
		record.ActualPath = actualFrame.OutputPath
		compare, err := ComparePNG(expectedFrame.OutputPath, actualFrame.OutputPath, opts)
		if err != nil {
			return FrameSetCompareReport{}, fmt.Errorf("aeoracle: compare frame %s: %w", expectedFrame.Tag, err)
		}
		record.Compare = &compare
		if compare.DifferentPixels > 0 {
			record.Status = FrameStatusDifferent
		} else {
			record.Status = FrameStatusOK
		}
		report.Frames = append(report.Frames, record)
	}

	extraTags := make([]string, 0)
	for tag := range actualFrames {
		if !seenActual[tag] {
			if _, expectedHasTag := expectedFrames[tag]; !expectedHasTag {
				extraTags = append(extraTags, tag)
			}
		}
	}
	sort.Strings(extraTags)
	for _, tag := range extraTags {
		actualFrame := actualFrames[tag]
		report.Frames = append(report.Frames, FrameCompareRecord{
			Tag:        actualFrame.Tag,
			Frame:      actualFrame.Frame,
			Seconds:    actualFrame.Seconds,
			Reason:     actualFrame.Reason,
			ActualPath: actualFrame.OutputPath,
			Status:     FrameStatusMissingExpected,
		})
	}

	report.Summary = summarizeFrameSet(report.Frames)
	return report, nil
}

func indexRenderedFrames(side string, frames []RenderedFrameRecord) (map[string]RenderedFrameRecord, error) {
	out := make(map[string]RenderedFrameRecord, len(frames))
	for _, frame := range frames {
		if frame.Tag == "" {
			return nil, fmt.Errorf("aeoracle: %s metadata contains frame with empty tag", side)
		}
		if _, exists := out[frame.Tag]; exists {
			return nil, fmt.Errorf("aeoracle: %s metadata contains duplicate frame tag %q", side, frame.Tag)
		}
		out[frame.Tag] = frame
	}
	return out, nil
}

func summarizeFrameSet(frames []FrameCompareRecord) FrameSetSummary {
	summary := FrameSetSummary{TotalFrames: len(frames)}
	for _, frame := range frames {
		switch frame.Status {
		case FrameStatusOK:
			summary.OKFrames++
		case FrameStatusDifferent:
			summary.DifferentFrames++
		case FrameStatusMissingExpected:
			summary.MissingExpectedFrames++
		case FrameStatusMissingActual:
			summary.MissingActualFrames++
		}
	}
	return summary
}
