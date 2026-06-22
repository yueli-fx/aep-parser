// flightdeck/showcase/booyah-clone/gen_text_komako.go — comp ② "テキスト変えるならココ！".
// One text layer "GLITCH" with a glitch-flicker layer transform (28-kf Position
// jitter + 24-kf Opacity flicker) and two text animators: Tracking Amount (2kf
// 135→0) + Character Offset (2kf 21→0). All values pulled from the original via
// the oracle; chunks built via our from-scratch text + animator + transform API.
//
// Two-phase: buildTextKomako creates the layer (NewTextLayer + SetText) before the
// project is reopened; finishTextKomako applies SetLayerTransform + the animators,
// which require a PARSED layer (see the Reopen in main). 29.97 fps (cloneFps),
// matching the original (the NewComposition fractional-fps cdta bug is fixed).
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const komakoCompName = "テキスト変えるならココ！"

func buildTextKomako(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, komakoCompName, 1920, 1080, cloneFps, 6)
	must(err)
	tl, err := aep.NewTextLayer(comp, "GLITCH")
	must(err)
	must(tl.SetText("GLITCH"))
}

// finishTextKomako runs after the project is reopened: it operates on the parsed
// "GLITCH" layer (SetLayerTransform + AddText*Animator require a parsed layer).
func finishTextKomako(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(komakoCompName)
	if comp == nil {
		panic("comp ②: not found after reopen")
	}
	layer := comp.LayerByName("GLITCH")
	if layer == nil {
		panic("comp ②: GLITCH layer not found after reopen")
	}

	// Match the original GLITCH text style (RE'd from its text doc via AE DOM:
	// Industry-Demi 110pt, base tracking 65, faux italic, centre-justified). The
	// font + centre justification + size are what let the original's anchor
	// (copied below) place the text dead-centre — without them a default-font
	// left-justified layout lands low. Industry-Demi need not be installed on this
	// box (AE substitutes at render); it resolves on a machine that has it.
	fontIdx, err := layer.AddFont("Industry-Demi")
	must(err)
	must(layer.SetRunFontIndex(0, fontIdx))
	must(layer.SetRunFontSize(0, 110))
	must(layer.SetRunTracking(0, 65))
	must(layer.SetRunFauxItalic(0, true))
	must(layer.SetParagraphJustification(0, aep.TextJustifyCenter))

	orig := orc.mustComp(komakoCompName).Layers[0]
	otg := findGroup(orig.PropertyTree(), "ADBE Transform Group")

	// --- layer transform: anchor (static) + Position (28kf) + Opacity (24kf) ---
	// The original anchors the text by its TEXT-BOX CENTRE (anchor Y -436 ==
	// its box centre, because its glyphs sit ~380px above the layer origin — a
	// btdk first-baseline our SetText doesn't reproduce). Our from-scratch text
	// is a normal point-text layout (baseline at the origin, box centre ≈ -27),
	// so copying the original's -436 anchor would push the text ~410px low.
	// Anchor by OUR OWN box centre instead, keeping the original's X; the
	// original Position keyframes then place that centre, matching the original.
	const cloneBoxCenterY = -27 // clone text-box centre (AE sourceRectAtTime, GLITCH 110pt)
	tr := aep.NewLayerTransform()
	anchorX := 8.98
	if ap := findProp(otg, "ADBE Anchor Point"); ap != nil && ap.StaticValue != nil {
		anchorX = toFloats(ap.StaticValue)[0]
	}
	must(tr.AnchorPoint().SetStaticValue([2]float64{anchorX, cloneBoxCenterY}))
	for _, kf := range findProp(otg, "ADBE Position").Keyframes {
		v := toFloats(kf.Value)
		must(tr.Position().AddKeyframeLinear(kf.Time, [2]float64{v[0], v[1]}))
	}
	// Parser reports Opacity normalized 0..1; SetLayerTransform's Opacity is percent.
	for _, kf := range findProp(otg, "ADBE Opacity").Keyframes {
		must(tr.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
	}
	must(aep.SetLayerTransform(layer, tr))

	// --- text animators: Tracking Amount + Character Offset (each 2kf, ease copied) ---
	trackKfs := scalarKfsOf(orig, "ADBE Text Tracking Amount")
	offKfs := scalarKfsOf(orig, "ADBE Text Character Offset")

	_, err = aep.AddTextTrackingAnimator(layer, trackKfs[0].Value, 0, 100, 0)
	must(err)
	must(aep.AnimateTextTracking(layer, 0, trackKfs))

	_, err = aep.AddTextCharacterOffsetAnimator(layer, offKfs[0].Value, 0, 100, 0)
	must(err)
	must(aep.AnimateTextCharacterOffset(layer, 0, offKfs))

	// Motion blur: the original's GLITCH layer has motion blur ON (the only layer
	// in the whole project that does). The fast Tracking/Character-Offset animation
	// then renders smeared — visible standalone AND through every comp that nests ②
	// (⑧ RGBズレ, ⑩). Both switches are required: the layer flag + the comp master
	// switch. Without them the clone's text is razor-sharp where the original blurs.
	must(layer.SetMotionBlur(true))
	must(comp.SetCompMotionBlur(true))

	fmt.Printf("  ② テキスト変えるならココ: GLITCH text + Position(%dkf)/Opacity(%dkf) + Tracking(%dkf)/CharOffset(%dkf)\n",
		len(findProp(otg, "ADBE Position").Keyframes), len(findProp(otg, "ADBE Opacity").Keyframes),
		len(trackKfs), len(offKfs))
}

// scalarKfsOf pulls a 1D-scalar animator leaf's keyframes (by match-name, anywhere
// under the layer's Text Animators) into []ScalarKeyframe, copying each keyframe's
// bezier temporal ease (In/Out Speed+Influence) verbatim from the original. The
// ease is NOT optional fidelity here: the original Tracking/Character Offset ease
// out hard (out-influence ≈0 at the first kf, in-influence ≈1 at the last) so the
// scramble resolves to "GLITCH" early and holds; a linear approximation keeps it
// mid-scramble far longer, which comp ⑧ (3 staggered copies of ②) magnified into
// three differently-scrambled RGB layers instead of a tight chromatic-aberration.
func scalarKfsOf(orig *aep.Layer, matchName string) []aep.ScalarKeyframe {
	tp := findGroup(orig.PropertyTree(), "ADBE Text Properties")
	animators := findGroup(tp, "ADBE Text Animators")
	var leaf *aep.Property
	for _, c := range animators.Children {
		ag, ok := c.(*aep.AEPropertyGroup)
		if !ok || ag.MatchName != "ADBE Text Animator" {
			continue
		}
		if p := findProp(findGroup(ag, "ADBE Text Animator Properties"), matchName); p != nil {
			leaf = p
			break
		}
	}
	if leaf == nil {
		panic("comp ②: animator leaf not found: " + matchName)
	}
	kfs := make([]aep.ScalarKeyframe, 0, len(leaf.Keyframes))
	for _, kf := range leaf.Keyframes {
		sk := aep.ScalarKeyframe{Time: kf.Time, Value: toScalar(kf.Value)}
		if len(kf.InTemporalEase) > 0 {
			sk.InEase = kf.InTemporalEase[0]
		}
		if len(kf.OutTemporalEase) > 0 {
			sk.OutEase = kf.OutTemporalEase[0]
		}
		kfs = append(kfs, sk)
	}
	return kfs
}
