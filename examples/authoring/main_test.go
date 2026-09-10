package main

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// Check the saved artifact, including edits made before structural deletion.
func TestReadmeProjectSurvivesWrite(t *testing.T) {
	var output bytes.Buffer
	if err := build().WriteAEP(&output); err != nil {
		t.Fatal(err)
	}
	project, err := aep.FromReader(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Compositions) != 1 {
		t.Fatalf("compositions: %d", len(project.Compositions))
	}
	comp := project.Compositions[0]
	if len(comp.Layers) != 2 || comp.LayerByName("Draft") != nil {
		t.Fatal("draft deletion did not survive writing")
	}
	title := comp.LayerByName("Title")
	if title == nil || title.TextSource == nil || title.TextSource.Text != "HELLO, AEP" {
		t.Fatal("text edit did not survive writing")
	}
	card := comp.LayerByName("Card")
	if card == nil || len(card.Effects) != 1 || card.Effects[0].MatchName != aep.EffectGaussianBlur {
		t.Fatal("Gaussian Blur did not survive writing")
	}
	for _, param := range card.Effects[0].Parameters {
		if param.MatchName != aep.EffectGaussianBlur+"-0001" {
			continue
		}
		if len(param.Keyframes) != 2 {
			t.Fatalf("blur keyframes: %d", len(param.Keyframes))
		}
		for i, want := range []aep.ScalarKeyframe{{Time: 0, Value: 30}, {Time: 1, Value: 0}} {
			got := param.Keyframes[i]
			if got.Time != want.Time || got.Value != want.Value {
				t.Fatalf("keyframe %d: time=%v value=%v, want %+v", i, got.Time, got.Value, want)
			}
		}
		return
	}
	t.Fatal("missing blur amount parameter")
}
