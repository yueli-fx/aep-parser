package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/projectindex"
)

func TestRunSearchEffectJSON(t *testing.T) {
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 7, Name: "Main", Layers: []*aep.Layer{
				{
					ID:    11,
					Index: 0,
					Name:  "Title",
					Effects: []*aep.Effect{
						{MatchName: "ADBE Glo2", Name: "Glow"},
					},
				},
			}},
		},
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runWithIO([]string{"search", "-aep", "fixture.aep", "-kind", "effect", "-query", "ADBE Glo2", "-json"}, stdout, stderr, fakeOpen(project))

	if code != 0 {
		t.Fatalf("run search = %d, want 0; stderr=%s", code, stderr.String())
	}
	var report SearchReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal search report: %v\n%s", err, stdout.String())
	}
	if report.Mode != "search" || report.ProjectPath != "fixture.aep" || report.Kind != "effect" || report.Query != "ADBE Glo2" {
		t.Fatalf("report header = %+v, want search/effect query", report)
	}
	if len(report.Hits) != 1 {
		t.Fatalf("len(Hits) = %d, want 1", len(report.Hits))
	}
	if report.Hits[0].Kind != projectindex.HitEffect || report.Hits[0].Location.LayerName != "Title" {
		t.Fatalf("hit = %+v, want effect hit on Title", report.Hits[0])
	}
}

func TestRunSearchReturnsOneForNoHits(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runWithIO([]string{"search", "-aep", "fixture.aep", "-kind", "effect", "-query", "Missing", "-json"}, stdout, stderr, fakeOpen(&aep.Project{}))

	if code != 1 {
		t.Fatalf("run search no hits = %d, want 1; stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
}

func TestRunCorpusJSON(t *testing.T) {
	project := &aep.Project{
		Compositions: []*aep.Composition{
			{ID: 7, Name: "Main", Layers: []*aep.Layer{
				{ID: 11, Index: 0, Name: "Title", Effects: []*aep.Effect{{MatchName: "ADBE Glo2"}}},
			}},
		},
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runWithIO([]string{"corpus", "-json", "one.aep", "two.aep"}, stdout, stderr, fakeOpen(project))

	if code != 0 {
		t.Fatalf("run corpus = %d, want 0; stderr=%s", code, stderr.String())
	}
	var corpus projectindex.Corpus
	if err := json.Unmarshal(stdout.Bytes(), &corpus); err != nil {
		t.Fatalf("unmarshal corpus: %v\n%s", err, stdout.String())
	}
	if len(corpus.Projects) != 2 {
		t.Fatalf("len(Projects) = %d, want 2", len(corpus.Projects))
	}
	if len(corpus.Facts) != 2 {
		t.Fatalf("len(Facts) = %d, want 2", len(corpus.Facts))
	}
	if corpus.Facts[0].Kind != projectindex.FactEffectUsage || corpus.Facts[0].Location.LayerName != "Title" {
		t.Fatalf("first fact = %+v, want effect usage on Title", corpus.Facts[0])
	}
	if strings.Contains(stdout.String(), "Pointers") || strings.Contains(stdout.String(), "pointers") {
		t.Fatalf("corpus JSON should not contain pointers:\n%s", stdout.String())
	}
}

func TestRunReturnsTwoForOpenError(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runWithIO([]string{"search", "-aep", "missing.aep", "-kind", "effect", "-query", "ADBE Glo2"}, stdout, stderr, func(string) (*aep.Project, error) {
		return nil, errors.New("boom")
	})

	if code != 2 {
		t.Fatalf("run open error = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "open missing.aep") {
		t.Fatalf("stderr = %q, want open path error", stderr.String())
	}
}

func fakeOpen(project *aep.Project) openProjectFunc {
	return func(string) (*aep.Project, error) {
		return project, nil
	}
}
