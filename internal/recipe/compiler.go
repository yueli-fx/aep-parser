package recipe

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/profile"
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
	if hasExpectedProfile(rec.ExpectedProfile) {
		prof, err := buildWrittenProfile(project, outPath)
		if err != nil {
			return report, fmt.Errorf("recipe: build profile for expected_profile: %w", err)
		}
		report.ProfileChecks = checkExpectedProfile(rec.ExpectedProfile, prof)
		for _, check := range report.ProfileChecks {
			if !check.Passed {
				report.Valid = false
				report.Refusals = append(report.Refusals, Refusal{
					Code:    "profile_contract_mismatch",
					Path:    check.Path,
					Message: check.Message,
				})
			}
		}
		if !report.Valid {
			return report, nil
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

func buildWrittenProfile(project *aep.Project, outPath string) (*profile.Profile, error) {
	var buf bytes.Buffer
	if err := project.WriteAEP(&buf); err != nil {
		return nil, err
	}
	reopened, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	return profile.Build(reopened, profile.Options{Path: outPath})
}

func hasExpectedProfile(expected ExpectedProfile) bool {
	return expected.CompCount != nil ||
		expected.LayerCount != nil ||
		expected.TextLayerCount != nil ||
		expected.ShapeLayerCount != nil ||
		len(expected.Effects) > 0 ||
		len(expected.Properties) > 0 ||
		len(expected.TextStyles) > 0 ||
		len(expected.Keyframes) > 0
}

func checkExpectedProfile(expected ExpectedProfile, prof *profile.Profile) []ProfileCheck {
	var checks []ProfileCheck
	add := func(path string, expected, actual any, passed bool) {
		msg := ""
		if !passed {
			msg = fmt.Sprintf("%s expected %v, got %v", path, expected, actual)
		}
		checks = append(checks, ProfileCheck{
			Path:     path,
			Passed:   passed,
			Expected: expected,
			Actual:   actual,
			Message:  msg,
		})
	}
	if expected.CompCount != nil {
		actual := prof.Fingerprint.CompCount
		add("expected_profile.comp_count", *expected.CompCount, actual, actual == *expected.CompCount)
	}
	if expected.LayerCount != nil {
		actual := prof.Fingerprint.LayerCount
		add("expected_profile.layer_count", *expected.LayerCount, actual, actual == *expected.LayerCount)
	}
	if expected.TextLayerCount != nil {
		actual := countProfileLayers(prof, func(layer profile.Layer) bool { return layer.Text != nil })
		add("expected_profile.text_layer_count", *expected.TextLayerCount, actual, actual == *expected.TextLayerCount)
	}
	if expected.ShapeLayerCount != nil {
		actual := countProfileLayers(prof, func(layer profile.Layer) bool { return len(layer.Shapes) > 0 })
		add("expected_profile.shape_layer_count", *expected.ShapeLayerCount, actual, actual == *expected.ShapeLayerCount)
	}
	for i, expectedEffect := range expected.Effects {
		effectPath := fmt.Sprintf("expected_profile.effects[%d]", i)
		effect := findProfileEffect(prof, expectedEffect.LayerName, expectedEffect.MatchName)
		add(effectPath, expectedEffect.MatchName, effectMatchName(effect), effect != nil)
		if effect == nil {
			continue
		}
		for pi, expectedParam := range expectedEffect.Params {
			paramPath := fmt.Sprintf("%s.params[%d]", effectPath, pi)
			param := findProfileParam(effect.Params, expectedParam.MatchName)
			if param == nil {
				add(paramPath, expectedParam.Value, nil, false)
				continue
			}
			passed := profileValueEqual(expectedParam.Value, param.StaticValue)
			add(paramPath, expectedParam.Value, param.StaticValue, passed)
		}
	}
	for i, expectedProp := range expected.Properties {
		propPath := fmt.Sprintf("expected_profile.properties[%d]", i)
		prop := findProfileLayerProperty(prof, expectedProp.LayerName, expectedProp.MatchName)
		if prop == nil {
			add(propPath, expectedProp.Value, nil, false)
			continue
		}
		passed := profileValueEqual(expectedProp.Value, prop.StaticValue)
		add(propPath, expectedProp.Value, prop.StaticValue, passed)
	}
	for i, expectedStyle := range expected.TextStyles {
		stylePath := fmt.Sprintf("expected_profile.text_styles[%d]", i)
		layer := findProfileLayer(prof, expectedStyle.LayerName)
		if layer == nil || layer.Text == nil {
			add(stylePath, "text layer", nil, false)
			continue
		}
		if expectedStyle.FontSize != nil {
			path := stylePath + ".font_size"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.FontSize
			})
			add(path, *expectedStyle.FontSize, actual, ok && math.Abs(actual-*expectedStyle.FontSize) < 1e-9)
		}
		if len(expectedStyle.FillColor) > 0 {
			path := stylePath + ".fill_color"
			actual, ok := profileRunColor(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) [4]float64 {
				return run.FillColor
			})
			add(path, expectedStyle.FillColor, actual, ok && profileValueEqual(expectedStyle.FillColor, actual))
		}
		if expectedStyle.Tracking != nil {
			path := stylePath + ".tracking"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.Tracking
			})
			add(path, *expectedStyle.Tracking, actual, ok && math.Abs(actual-*expectedStyle.Tracking) < 1e-9)
		}
		if expectedStyle.FauxBold != nil {
			path := stylePath + ".faux_bold"
			actual, ok := profileRunBool(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) bool {
				return run.FauxBold
			})
			add(path, *expectedStyle.FauxBold, actual, ok && actual == *expectedStyle.FauxBold)
		}
		if expectedStyle.FauxItalic != nil {
			path := stylePath + ".faux_italic"
			actual, ok := profileRunBool(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) bool {
				return run.FauxItalic
			})
			add(path, *expectedStyle.FauxItalic, actual, ok && actual == *expectedStyle.FauxItalic)
		}
		if expectedStyle.ApplyStroke != nil {
			path := stylePath + ".apply_stroke"
			actual, ok := profileRunBool(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) bool {
				return run.ApplyStroke
			})
			add(path, *expectedStyle.ApplyStroke, actual, ok && actual == *expectedStyle.ApplyStroke)
		}
		if len(expectedStyle.StrokeColor) > 0 {
			path := stylePath + ".stroke_color"
			actual, ok := profileRunColor(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) [4]float64 {
				return run.StrokeColor
			})
			add(path, expectedStyle.StrokeColor, actual, ok && profileValueEqual(expectedStyle.StrokeColor, actual))
		}
		if expectedStyle.StrokeWidth != nil {
			path := stylePath + ".stroke_width"
			actual, ok := profileRunFloat(layer.Text.Runs, expectedStyle.RunIndex, func(run profile.TextStyleRun) float64 {
				return run.StrokeWidth
			})
			add(path, *expectedStyle.StrokeWidth, actual, ok && math.Abs(actual-*expectedStyle.StrokeWidth) < 1e-9)
		}
		if expectedStyle.Justification != "" {
			path := stylePath + ".justification"
			actual, ok := profileParagraphJustification(layer.Text.Paragraphs, expectedStyle.ParagraphIndex)
			add(path, expectedStyle.Justification, actual, ok && strings.EqualFold(actual, expectedStyle.Justification))
		}
	}
	for i, expectedKeyframes := range expected.Keyframes {
		kfPropPath := fmt.Sprintf("expected_profile.keyframes[%d]", i)
		prop := findProfileLayerProperty(prof, expectedKeyframes.LayerName, expectedKeyframes.MatchName)
		if prop == nil {
			add(kfPropPath, expectedKeyframes.MatchName, nil, false)
			continue
		}
		add(kfPropPath+".count", len(expectedKeyframes.Keyframes), len(prop.Keyframes), len(prop.Keyframes) == len(expectedKeyframes.Keyframes))
		for ki, expectedKF := range expectedKeyframes.Keyframes {
			kfPath := fmt.Sprintf("%s.keyframes[%d]", kfPropPath, ki)
			if ki >= len(prop.Keyframes) {
				add(kfPath, expectedKF.Value, nil, false)
				continue
			}
			actualKF := prop.Keyframes[ki]
			timeOK := math.Abs(actualKF.Time-expectedKF.Time) < 1e-6
			valueOK := profileValueEqual(expectedKF.Value, actualKF.Value)
			add(kfPath, expectedKF.Value, actualKF.Value, timeOK && valueOK)
			if !timeOK {
				add(kfPath+".time", expectedKF.Time, actualKF.Time, false)
			}
		}
	}
	return checks
}

func countProfileLayers(prof *profile.Profile, include func(profile.Layer) bool) int {
	var count int
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if include(layer) {
				count++
			}
		}
	}
	return count
}

func findProfileEffect(prof *profile.Profile, layerName, matchName string) *profile.Effect {
	layer := findProfileLayer(prof, layerName)
	if layer == nil {
		return nil
	}
	for i := range layer.Effects {
		if layer.Effects[i].MatchName == matchName {
			return &layer.Effects[i]
		}
	}
	return nil
}

func findProfileLayer(prof *profile.Profile, layerName string) *profile.Layer {
	for _, comp := range prof.Comps {
		for i := range comp.Layers {
			if comp.Layers[i].Name == layerName {
				return &comp.Layers[i]
			}
		}
	}
	return nil
}

func effectMatchName(effect *profile.Effect) any {
	if effect == nil {
		return nil
	}
	return effect.MatchName
}

func findProfileParam(params []profile.Property, matchName string) *profile.Property {
	for i := range params {
		if params[i].MatchName == matchName {
			return &params[i]
		}
	}
	return nil
}

func findProfileLayerProperty(prof *profile.Profile, layerName, matchName string) *profile.Property {
	layer := findProfileLayer(prof, layerName)
	if layer == nil {
		return nil
	}
	if prop := findProfileParam(layer.Properties, matchName); prop != nil {
		return prop
	}
	for i := range layer.Shapes {
		if prop := findProfileParam(layer.Shapes[i].Properties, matchName); prop != nil {
			return prop
		}
	}
	return nil
}

func profileRunFloat(runs []profile.TextStyleRun, index int, value func(profile.TextStyleRun) float64) (float64, bool) {
	if index < 0 || index >= len(runs) {
		return 0, false
	}
	return value(runs[index]), true
}

func profileRunBool(runs []profile.TextStyleRun, index int, value func(profile.TextStyleRun) bool) (bool, bool) {
	if index < 0 || index >= len(runs) {
		return false, false
	}
	return value(runs[index]), true
}

func profileRunColor(runs []profile.TextStyleRun, index int, value func(profile.TextStyleRun) [4]float64) ([]float64, bool) {
	if index < 0 || index >= len(runs) {
		return nil, false
	}
	color := value(runs[index])
	return []float64{color[0], color[1], color[2], color[3]}, true
}

func profileParagraphJustification(paragraphs []profile.TextParagraph, index int) (string, bool) {
	if index < 0 || index >= len(paragraphs) {
		return "", false
	}
	return paragraphs[index].Justification, true
}

func profileValueEqual(expected, actual any) bool {
	expected, err := normalizeEffectParamValue(expected)
	if err != nil {
		return false
	}
	actual, err = normalizeEffectParamValue(actual)
	if err != nil {
		return false
	}
	switch e := expected.(type) {
	case float64:
		a, ok := actual.(float64)
		return ok && math.Abs(e-a) < 1e-9
	case []float64:
		a, ok := actual.([]float64)
		if !ok || len(e) != len(a) {
			return false
		}
		for i := range e {
			if math.Abs(e[i]-a[i]) >= 1e-9 {
				return false
			}
		}
		return true
	default:
		return false
	}
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
			fx, err := aep.AddEffect(layer, effect.MatchName)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q add effect %q: %w", layerSpec.Name, effect.MatchName, err)
			}
			for _, param := range effect.Params {
				value, err := normalizeEffectParamValue(param.Value)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q effect %q param %q: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
				}
				if _, err := aep.SetEffectParam(layer, fx, param.MatchName, value); err != nil {
					return nil, fmt.Errorf("recipe: layer %q effect %q param %q: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
				}
			}
		}
	}
	return reopened, nil
}

func normalizeEffectParamValue(value any) (any, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case []float64:
		if len(v) == 0 {
			return nil, fmt.Errorf("empty numeric array")
		}
		return v, nil
	case []any:
		if len(v) == 0 {
			return nil, fmt.Errorf("empty numeric array")
		}
		out := make([]float64, 0, len(v))
		for i, item := range v {
			n, ok := item.(float64)
			if !ok {
				return nil, fmt.Errorf("array item %d is %T, want number", i, item)
			}
			out = append(out, n)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
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
		if spec.TextStyle != nil {
			if err := applyTextStyle(l, *spec.TextStyle); err != nil {
				return fmt.Errorf("recipe: text layer %q style: %w", spec.Name, err)
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

func applyTextStyle(layer *aep.Layer, spec TextStyleSpec) error {
	if spec.FontSize != nil {
		if err := layer.SetRunFontSize(spec.RunIndex, *spec.FontSize); err != nil {
			return err
		}
	}
	if len(spec.FillColor) >= 3 {
		if err := layer.SetRunFillColor(spec.RunIndex, rgbaColor(spec.FillColor)); err != nil {
			return err
		}
	}
	if spec.Tracking != nil {
		if err := layer.SetRunTracking(spec.RunIndex, *spec.Tracking); err != nil {
			return err
		}
	}
	if spec.FauxBold != nil {
		if err := layer.SetRunFauxBold(spec.RunIndex, *spec.FauxBold); err != nil {
			return err
		}
	}
	if spec.FauxItalic != nil {
		if err := layer.SetRunFauxItalic(spec.RunIndex, *spec.FauxItalic); err != nil {
			return err
		}
	}
	if spec.ApplyStroke != nil {
		if err := layer.SetRunApplyStroke(spec.RunIndex, *spec.ApplyStroke); err != nil {
			return err
		}
	}
	if len(spec.StrokeColor) >= 3 {
		if err := layer.SetRunStrokeColor(spec.RunIndex, rgbaColor(spec.StrokeColor)); err != nil {
			return err
		}
	}
	if spec.StrokeWidth != nil {
		if err := layer.SetRunStrokeWidth(spec.RunIndex, *spec.StrokeWidth); err != nil {
			return err
		}
	}
	if spec.Justification != "" {
		justification, err := textJustification(spec.Justification)
		if err != nil {
			return err
		}
		if err := layer.SetParagraphJustification(spec.ParagraphIndex, justification); err != nil {
			return err
		}
	}
	return nil
}

func textJustification(value string) (aep.TextJustification, error) {
	switch value {
	case "left":
		return aep.TextJustifyLeft, nil
	case "right":
		return aep.TextJustifyRight, nil
	case "center":
		return aep.TextJustifyCenter, nil
	default:
		return 0, fmt.Errorf("unsupported justification %q", value)
	}
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
		if len(shape.Position) == 2 {
			if err := rect.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
		if shape.Roundness != nil {
			if err := rect.SetRoundness(*shape.Roundness); err != nil {
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
		if len(shape.Position) == 2 {
			if err := ellipse.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
	}
	if shape.RoundCorners != nil {
		roundCorners, err := group.AddRoundCorners()
		if err != nil {
			return err
		}
		if shape.RoundCorners.Radius != nil {
			if err := roundCorners.SetRadius(*shape.RoundCorners.Radius); err != nil {
				return err
			}
		}
	}
	if shape.OffsetPaths != nil {
		offsetPaths, err := group.AddOffsetPaths()
		if err != nil {
			return err
		}
		if shape.OffsetPaths.Amount != nil {
			if err := offsetPaths.SetAmount(*shape.OffsetPaths.Amount); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.LineJoin != "" {
			lineJoin, err := offsetLineJoin(shape.OffsetPaths.LineJoin)
			if err != nil {
				return err
			}
			if err := offsetPaths.SetLineJoin(lineJoin); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.MiterLimit != nil {
			if err := offsetPaths.SetMiterLimit(*shape.OffsetPaths.MiterLimit); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.Copies != nil {
			if err := offsetPaths.SetCopies(*shape.OffsetPaths.Copies); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.CopyOffset != nil {
			if err := offsetPaths.SetCopyOffset(*shape.OffsetPaths.CopyOffset); err != nil {
				return err
			}
		}
	}
	if shape.ZigZag != nil {
		zigZag, err := group.AddZigZag()
		if err != nil {
			return err
		}
		if shape.ZigZag.Size != nil {
			if err := zigZag.SetSize(*shape.ZigZag.Size); err != nil {
				return err
			}
		}
		if shape.ZigZag.Detail != nil {
			if err := zigZag.SetDetail(*shape.ZigZag.Detail); err != nil {
				return err
			}
		}
		if shape.ZigZag.Points != "" {
			points, err := zigZagPoints(shape.ZigZag.Points)
			if err != nil {
				return err
			}
			if err := zigZag.SetPoints(points); err != nil {
				return err
			}
		}
	}
	if shape.Trim != nil {
		trim, err := group.AddTrim()
		if err != nil {
			return err
		}
		if shape.Trim.Start != nil {
			if err := trim.SetStart(*shape.Trim.Start); err != nil {
				return err
			}
		}
		if shape.Trim.End != nil {
			if err := trim.SetEnd(*shape.Trim.End); err != nil {
				return err
			}
		}
		if shape.Trim.Offset != nil {
			if err := trim.SetOffset(*shape.Trim.Offset); err != nil {
				return err
			}
		}
	}
	if len(shape.FillColor) >= 3 || shape.FillOpacity != nil {
		fill, err := group.AddFill()
		if err != nil {
			return err
		}
		if len(shape.FillColor) >= 3 {
			if err := fill.SetColor(rgbaColor(shape.FillColor)); err != nil {
				return err
			}
		}
		if shape.FillOpacity != nil {
			if err := fill.SetOpacity(*shape.FillOpacity); err != nil {
				return err
			}
		}
	}
	if shape.Stroke != nil {
		stroke, err := group.AddStroke()
		if err != nil {
			return err
		}
		if len(shape.Stroke.Color) >= 3 {
			color := rgbaColor(shape.Stroke.Color)
			if err := stroke.SetColor(color); err != nil {
				return err
			}
		}
		if shape.Stroke.Width != nil {
			if err := stroke.SetWidth(*shape.Stroke.Width); err != nil {
				return err
			}
		}
		if shape.Stroke.Opacity != nil {
			if err := stroke.SetOpacity(*shape.Stroke.Opacity); err != nil {
				return err
			}
		}
	}
	return nil
}

func offsetLineJoin(value string) (aep.StrokeLineJoin, error) {
	switch value {
	case "miter":
		return aep.StrokeLineJoinMiter, nil
	case "round":
		return aep.StrokeLineJoinRound, nil
	case "bevel":
		return aep.StrokeLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported offset line_join %q", value)
	}
}

func zigZagPoints(value string) (aep.ZigZagPoints, error) {
	switch value {
	case "corner":
		return aep.ZigZagPointsCorner, nil
	case "smooth":
		return aep.ZigZagPointsSmooth, nil
	default:
		return 0, fmt.Errorf("unsupported zigzag points %q", value)
	}
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
	for _, kf := range spec.AnchorPointKeyframes {
		if err := t.AnchorPoint().AddKeyframeLinear(kf.Time, [2]float64{kf.Value[0], kf.Value[1]}); err != nil {
			return err
		}
	}
	for _, kf := range spec.ScaleKeyframes {
		if err := t.Scale().AddKeyframeLinear(kf.Time, [2]float64{kf.Value[0], kf.Value[1]}); err != nil {
			return err
		}
	}
	for _, kf := range spec.RotationKeyframes {
		if err := t.Rotation().AddKeyframeLinear(kf.Time, kf.Value); err != nil {
			return err
		}
	}
	for _, kf := range spec.OpacityKeyframes {
		if err := t.Opacity().AddKeyframeLinear(kf.Time, kf.Value); err != nil {
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

func rgbaColor(values []float64) [4]float64 {
	alpha := 1.0
	if len(values) >= 4 {
		alpha = toUnitColor(values[3])
	}
	return [4]float64{
		toUnitColor(values[0]),
		toUnitColor(values[1]),
		toUnitColor(values[2]),
		alpha,
	}
}
