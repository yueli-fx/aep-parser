package selfhost

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

type SampleShellEffectTemplateAuditOptions struct {
	CandidatesDir string
	OutPath       string
}

type SampleShellEffectTemplateAuditResult struct {
	SchemaVersion int                                   `json:"schema_version"`
	CandidatesDir string                                `json:"candidates_dir"`
	ResultPath    string                                `json:"result_path,omitempty"`
	Summary       SampleShellEffectTemplateAuditSummary `json:"summary"`
	Candidates    []SampleShellEffectTemplateAuditRow   `json:"candidates"`
}

type SampleShellEffectTemplateAuditSummary struct {
	Total      int `json:"total"`
	Candidate  int `json:"candidate"`
	Review     int `json:"review"`
	Reject     int `json:"reject"`
	ParseError int `json:"parse_error"`
}

type SampleShellEffectTemplateAuditRow struct {
	MatchName        string `json:"match_name,omitempty"`
	Path             string `json:"path"`
	Bytes            int    `json:"bytes,omitempty"`
	WrapperForm      string `json:"wrapper_form,omitempty"`
	ParamValueGroups int    `json:"param_value_groups,omitempty"`
	CdatValues       int    `json:"cdat_values,omitempty"`
	KeyframeValues   int    `json:"keyframe_values,omitempty"`
	Expressions      int    `json:"expressions,omitempty"`
	LayerRefs        int    `json:"layer_refs,omitempty"`
	PardDefs         int    `json:"pard_defs,omitempty"`
	Recommendation   string `json:"recommendation"`
	Error            string `json:"error,omitempty"`
}

func AuditSampleShellEffectTemplateCandidates(opts SampleShellEffectTemplateAuditOptions) (SampleShellEffectTemplateAuditResult, error) {
	if opts.CandidatesDir == "" {
		opts.CandidatesDir = filepath.Join("tmp", "effect_template_candidates")
	}
	if opts.OutPath == "" {
		opts.OutPath = filepath.Join(opts.CandidatesDir, "candidate_audit.json")
	}
	result := SampleShellEffectTemplateAuditResult{
		SchemaVersion: 1,
		CandidatesDir: opts.CandidatesDir,
		ResultPath:    opts.OutPath,
		Candidates:    []SampleShellEffectTemplateAuditRow{},
	}
	paths, err := sampleShellEffectTemplateCandidatePaths(opts.CandidatesDir)
	if err != nil {
		return SampleShellEffectTemplateAuditResult{}, err
	}
	for _, path := range paths {
		row := auditSampleShellEffectTemplateCandidate(path)
		result.Candidates = append(result.Candidates, row)
		result.Summary.Total++
		switch row.Recommendation {
		case "candidate_clean":
			result.Summary.Candidate++
		case "review_materialized_defaults":
			result.Summary.Review++
		case "reject_candidate":
			result.Summary.Reject++
		case "parse_error":
			result.Summary.ParseError++
		}
	}
	if err := writeIndentedJSON(opts.OutPath, result); err != nil {
		return SampleShellEffectTemplateAuditResult{}, err
	}
	return result, nil
}

func sampleShellEffectTemplateCandidatePaths(dir string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".bin" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func auditSampleShellEffectTemplateCandidate(path string) SampleShellEffectTemplateAuditRow {
	data, err := os.ReadFile(path)
	if err != nil {
		return SampleShellEffectTemplateAuditRow{Path: path, Recommendation: "parse_error", Error: err.Error()}
	}
	row := SampleShellEffectTemplateAuditRow{Path: path, Bytes: len(data)}
	wrapper, err := rifx.ReadChunk(bytes.NewReader(data))
	if err != nil {
		row.Recommendation = "parse_error"
		row.Error = err.Error()
		return row
	}
	row.WrapperForm = wrapper.FormType.String()
	if len(wrapper.Children) < 2 || wrapper.Children[0].ID != rifx.IDTdmn || !wrapper.Children[1].IsList() {
		row.Recommendation = "parse_error"
		row.Error = fmt.Sprintf("unexpected template wrapper shape: children=%d", len(wrapper.Children))
		return row
	}
	row.MatchName = sampleShellTrimNUL(wrapper.Children[0].Data)
	auditSampleShellEffectTemplateChunk(wrapper.Children[1], false, &row)
	row.Recommendation = sampleShellEffectTemplateRecommendation(row)
	return row
}

func auditSampleShellEffectTemplateChunk(chunk *rifx.Chunk, inTdbs bool, row *SampleShellEffectTemplateAuditRow) {
	if chunk == nil {
		return
	}
	nowInTdbs := inTdbs
	if chunk.IsList() && chunk.FormType == rifx.IDTdbs {
		row.ParamValueGroups++
		nowInTdbs = true
	}
	if chunk.ID == rifx.IDCdat {
		row.CdatValues++
	}
	if chunk.IsList() && chunk.FormType == rifx.IDkfl {
		row.KeyframeValues++
	}
	if nowInTdbs && chunk.ID == rifx.IDUtf8 {
		row.Expressions++
	}
	if chunk.ID == rifx.IDTdpi {
		row.LayerRefs++
	}
	if chunk.ID == rifx.IDpard {
		row.PardDefs++
	}
	for _, child := range chunk.Children {
		auditSampleShellEffectTemplateChunk(child, nowInTdbs, row)
	}
}

func sampleShellEffectTemplateRecommendation(row SampleShellEffectTemplateAuditRow) string {
	if row.KeyframeValues > 0 || row.Expressions > 0 {
		return "reject_candidate"
	}
	if row.CdatValues > 0 || row.LayerRefs > 0 {
		return "review_materialized_defaults"
	}
	return "candidate_clean"
}
