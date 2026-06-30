package serializer

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/scene"
)

func TestParseWarningSinkRecordsStructuredProjectWarning(t *testing.T) {
	p := &Project{}
	ctx := newParseCtxWithWarningSink(0, "Main", projectWarningSink(p))

	ctx.warn("chunk %s is short", "lhd3")

	if got, want := len(p.Warnings), 1; got != want {
		t.Fatalf("Warnings len = %d, want %d", got, want)
	}
	if got, want := len(p.ParseWarnings), 1; got != want {
		t.Fatalf("ParseWarnings len = %d, want %d", got, want)
	}
	wantMessage := `comp "Main": chunk lhd3 is short`
	if p.Warnings[0] != wantMessage {
		t.Fatalf("Warnings[0] = %q, want %q", p.Warnings[0], wantMessage)
	}
	if got := p.ParseWarnings[0]; got != (scene.ParseWarning{Message: wantMessage}) {
		t.Fatalf("ParseWarnings[0] = %+v, want message %q", got, wantMessage)
	}
}
