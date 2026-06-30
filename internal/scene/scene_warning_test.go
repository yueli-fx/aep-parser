package scene

import "testing"

func TestProjectRecordParseWarningKeepsStringAndStructuredWarningsInSync(t *testing.T) {
	p := &Project{}

	p.recordParseWarning(ParseWarning{
		Chunk:   "lhd3",
		Offset:  42,
		Message: "short keyframe easing stream",
	})

	if got, want := len(p.Warnings), 1; got != want {
		t.Fatalf("Warnings len = %d, want %d", got, want)
	}
	if got, want := len(p.ParseWarnings), 1; got != want {
		t.Fatalf("ParseWarnings len = %d, want %d", got, want)
	}
	if p.Warnings[0] != "short keyframe easing stream" {
		t.Fatalf("Warnings[0] = %q", p.Warnings[0])
	}
	if got := p.ParseWarnings[0]; got.Chunk != "lhd3" || got.Offset != 42 || got.Message != "short keyframe easing stream" {
		t.Fatalf("ParseWarnings[0] = %+v", got)
	}
}

func TestProjectRollbackParseWarningsTruncatesBothViews(t *testing.T) {
	p := &Project{}
	p.recordParseWarning(ParseWarning{Message: "before"})
	warningsLen, parseWarningsLen := p.warningLengths()

	p.recordParseWarning(ParseWarning{Message: "after"})
	p.rollbackWarnings(warningsLen, parseWarningsLen)

	if got, want := p.Warnings, []string{"before"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("Warnings after rollback = %v, want %v", got, want)
	}
	if got, want := p.ParseWarnings, []ParseWarning{{Message: "before"}}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("ParseWarnings after rollback = %+v, want %+v", got, want)
	}
}
