package showcase_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestRainShowcaseGeneratorBuildsNativeRainSpine(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	out := filepath.Join(repoRoot, "flightdeck", "showcase", "rain", "rain.aep")
	_ = os.Remove(out)

	cmd := exec.Command("go", "run", "./flightdeck/showcase/rain")
	cmd.Dir = repoRoot
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run rain showcase failed: %v\n%s", err, b)
	}

	p, err := aep.Open(out)
	if err != nil {
		t.Fatalf("open generated rain.aep: %v", err)
	}

	if p.CompositionByName("RAIN") == nil {
		t.Fatalf("generated project missing RAIN comp")
	}
	if p.CompositionByName("RAIN_STREAKS") == nil {
		t.Fatalf("generated project missing RAIN_STREAKS precomp")
	}

	rain := p.CompositionByName("RAIN")
	requireLayer(t, rain, "RainStreaks")
	requireLayer(t, rain, "NativeMist")
	requireLayer(t, rain, "EchoTrail")
	requireLayer(t, rain, "BG")

	streaks := p.CompositionByName("RAIN_STREAKS")
	if len(streaks.Layers) < 4 {
		t.Fatalf("RAIN_STREAKS layers = %d, want at least 4 native streak layers", len(streaks.Layers))
	}
}

func requireLayer(t *testing.T, c *aep.Composition, name string) {
	t.Helper()
	if c.LayerByName(name) == nil {
		t.Fatalf("%s missing layer %q", c.Name, name)
	}
}
