package aep_test

import (
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestInsertLayer_RefuseNilSrc(t *testing.T) {
	c := &aep.Composition{Layers: []*aep.Layer{}}
	_, err := c.InsertLayer(nil, 0)
	if err == nil {
		t.Fatal("expected refuse on nil src, got nil error")
	}
	if !strings.Contains(err.Error(), "src cannot be nil") {
		t.Errorf("error should mention 'src cannot be nil', got %q", err.Error())
	}
}
