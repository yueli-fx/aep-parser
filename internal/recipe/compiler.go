package recipe

import (
	"fmt"
	"os"
	"path/filepath"

	aep "github.com/example/aep-parser/internal/aep"
)

func CompileToFile(rec Recipe, outPath string, caps CapabilityIndex) (Report, error) {
	report := ValidateWithCapabilities(rec, caps)
	report.OutputPath = outPath
	if !report.Valid {
		return report, nil
	}
	if outPath == "" {
		report.Valid = false
		report.Refusals = append(report.Refusals, Refusal{Code: "missing_output_path", Path: "output_path", Message: "output path is required"})
		return report, nil
	}

	project := aep.NewProject(aep.TargetAE2020)
	compSpec := rec.Comps[0]
	comp, err := aep.NewComposition(project, compSpec.Name, uint16(compSpec.Width), uint16(compSpec.Height), compSpec.FrameRate, compSpec.Duration)
	if err != nil {
		return report, fmt.Errorf("recipe: create comp: %w", err)
	}
	for _, layerSpec := range compSpec.Layers {
		if err := compileLayer(comp, layerSpec, compSpec); err != nil {
			return report, err
		}
	}
	if hasEffects(compSpec) {
		project, err = materializeEffects(project, compSpec)
		if err != nil {
			return report, err
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return report, err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return report, err
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		return report, err
	}
	return report, nil
}

func hasEffects(comp CompSpec) bool {
	for _, layer := range comp.Layers {
		if len(layer.Effects) > 0 {
			return true
		}
	}
	return false
}

func materializeEffects(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("recipe: reopen for effects: %w", err)
	}
	if len(reopened.Compositions) == 0 {
		return nil, fmt.Errorf("recipe: reopen for effects: no compositions")
	}
	comp := reopened.Compositions[0]
	for i, layerSpec := range compSpec.Layers {
		if len(layerSpec.Effects) == 0 {
			continue
		}
		if i >= len(comp.Layers) {
			return nil, fmt.Errorf("recipe: reopen for effects: layer index %d missing", i)
		}
		layer := comp.Layers[i]
		for _, effect := range layerSpec.Effects {
			if _, err := aep.AddEffect(layer, effect.MatchName); err != nil {
				return nil, fmt.Errorf("recipe: layer %q add effect %q: %w", layerSpec.Name, effect.MatchName, err)
			}
		}
	}
	return reopened, nil
}

func compileLayer(comp *aep.Composition, spec Layer, compSpec CompSpec) error {
	var layer *aep.Layer
	switch spec.Type {
	case "text":
		l, err := aep.NewTextLayer(comp, spec.Name)
		if err != nil {
			return fmt.Errorf("recipe: text layer %q: %w", spec.Name, err)
		}
		if spec.Text != "" {
			if err := l.SetText(spec.Text); err != nil {
				return fmt.Errorf("recipe: text layer %q set text: %w", spec.Name, err)
			}
		}
		layer = l
	case "shape":
		l, err := aep.NewShapeLayer(comp, spec.Name)
		if err != nil {
			return fmt.Errorf("recipe: shape layer %q: %w", spec.Name, err)
		}
		if spec.Shape != nil {
			if err := compileShape(l.RootGroup(), *spec.Shape); err != nil {
				return fmt.Errorf("recipe: shape layer %q: %w", spec.Name, err)
			}
		}
		layer = l.Layer
	case "solid":
		color := [3]float64{0, 0, 0}
		if spec.Shape != nil && len(spec.Shape.FillColor) >= 3 {
			color = [3]float64{toUnitColor(spec.Shape.FillColor[0]), toUnitColor(spec.Shape.FillColor[1]), toUnitColor(spec.Shape.FillColor[2])}
		}
		l, err := aep.NewSolidLayer(comp, spec.Name, compSpec.Width, compSpec.Height, color)
		if err != nil {
			return fmt.Errorf("recipe: solid layer %q: %w", spec.Name, err)
		}
		layer = l
	default:
		return fmt.Errorf("recipe: unsupported layer type %q", spec.Type)
	}
	if layer != nil {
		if err := applyTransform(layer, spec.Transform); err != nil {
			return fmt.Errorf("recipe: layer %q transform: %w", spec.Name, err)
		}
	}
	return nil
}

func compileShape(group *aep.VectorGroup, shape ShapeSpec) error {
	switch shape.Kind {
	case "rect":
		rect, err := group.AddRect()
		if err != nil {
			return err
		}
		if len(shape.Size) == 2 {
			if err := rect.SetSize([2]float64{shape.Size[0], shape.Size[1]}); err != nil {
				return err
			}
		}
	case "ellipse":
		ellipse, err := group.AddEllipse()
		if err != nil {
			return err
		}
		if len(shape.Size) == 2 {
			if err := ellipse.SetSize([2]float64{shape.Size[0], shape.Size[1]}); err != nil {
				return err
			}
		}
	}
	if len(shape.FillColor) >= 3 {
		fill, err := group.AddFill()
		if err != nil {
			return err
		}
		alpha := 1.0
		if len(shape.FillColor) >= 4 {
			alpha = toUnitColor(shape.FillColor[3])
		}
		if err := fill.SetColor([4]float64{
			toUnitColor(shape.FillColor[0]),
			toUnitColor(shape.FillColor[1]),
			toUnitColor(shape.FillColor[2]),
			alpha,
		}); err != nil {
			return err
		}
	}
	return nil
}

func applyTransform(layer *aep.Layer, spec Transform) error {
	t := aep.NewLayerTransform()
	if len(spec.AnchorPoint) == 2 {
		if err := t.AnchorPoint().SetStaticValue([2]float64{spec.AnchorPoint[0], spec.AnchorPoint[1]}); err != nil {
			return err
		}
	}
	if len(spec.Position) == 2 {
		if err := t.Position().SetStaticValue([2]float64{spec.Position[0], spec.Position[1]}); err != nil {
			return err
		}
	}
	if len(spec.Scale) == 2 {
		if err := t.Scale().SetStaticValue([2]float64{spec.Scale[0], spec.Scale[1]}); err != nil {
			return err
		}
	}
	if spec.Rotation != nil {
		if err := t.Rotation().SetStaticValue(*spec.Rotation); err != nil {
			return err
		}
	}
	if spec.Opacity != nil {
		if err := t.Opacity().SetStaticValue(*spec.Opacity); err != nil {
			return err
		}
	}
	for _, kf := range spec.PositionKeyframes {
		if err := t.Position().AddKeyframeLinear(kf.Time, [2]float64{kf.Value[0], kf.Value[1]}); err != nil {
			return err
		}
	}
	return aep.SetLayerTransform(layer, t)
}

func toUnitColor(v float64) float64 {
	if v > 1 {
		return v / 255
	}
	return v
}
