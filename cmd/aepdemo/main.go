// cmd/aepdemo —— 用 aep-parser 现有"创建"能力从零生成一个 .aep 功能演示文件。
//
// 一个合成里放多个形状图层，每个图层名即它演示的功能。打开 AE 后展开该合成
// 逐层查看。（多图层同合成此前会被 AE silent-drop，已由 lower_layer_siblings
// 修复 —— 见 incident-reports/multi-layer-silent-drop.md。）
//
//	go run ./cmd/aepdemo                 # 默认写到 ./demo_features.aep
//	go run ./cmd/aepdemo out.aep         # 指定输出路径
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

var (
	red   = [4]float64{0.90, 0.22, 0.22, 1}
	blue  = [4]float64{0.25, 0.45, 1.00, 1}
	green = [4]float64{0.25, 0.78, 0.36, 1}
	white = [4]float64{1, 1, 1, 1}
)

func main() {
	out := "demo_features.aep"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	// 能力①：从零创建项目（目标 AE 2025）
	p := aep.NewProject(aep.TargetAE2025)
	// 能力②：从零创建合成
	comp, err := p.NewComposition("功能演示 aep-parser", 1920, 1080, 30, 6)
	must(err)

	// 能力③：从零创建形状图层 + 取名（同一合成内可放多层）
	layer := func(name string, at [2]float64) *aep.ShapeLayer {
		l, err := comp.NewShapeLayer(name)
		must(err)
		must(l.Position().SetStaticValue(at))
		return l
	}

	// 矩形 + 纯色填充
	{
		l := layer("① 矩形 AddRect + 纯色填充 AddFill(红)", [2]float64{320, 220})
		r, _ := l.RootGroup().AddRect()
		must(r.SetSize([2]float64{300, 190}))
		f, _ := l.RootGroup().AddFill()
		must(f.SetColor(red))
	}
	// 椭圆 + 描边
	{
		l := layer("② 椭圆 AddEllipse + 描边 AddStroke(蓝,宽10)", [2]float64{780, 220})
		e, _ := l.RootGroup().AddEllipse()
		must(e.SetSize([2]float64{220, 220}))
		s, _ := l.RootGroup().AddStroke()
		must(s.SetColor(blue))
		must(s.SetWidth(10))
	}
	// 自定义路径（闭合五边形）+ 填充
	{
		l := layer("③ 自定义路径 AddPath(五边形) + 绿填充", [2]float64{1240, 220})
		pa, _ := l.RootGroup().AddPath()
		must(pa.SetVertices([][2]float64{{0, -110}, {105, -34}, {65, 89}, {-65, 89}, {-105, -34}}))
		must(pa.SetClosed(true))
		f, _ := l.RootGroup().AddFill()
		must(f.SetColor(green))
	}
	// 渐变填充
	{
		l := layer("④ 渐变填充 AddGradientFill(黑→白)", [2]float64{1680, 220})
		r, _ := l.RootGroup().AddRect()
		must(r.SetSize([2]float64{300, 190}))
		_, _ = l.RootGroup().AddGradientFill()
	}
	// 描边样式枚举
	{
		l := layer("⑤ 描边样式 SetLineCap/LineJoin + 圆角矩形", [2]float64{420, 560})
		r, _ := l.RootGroup().AddRect()
		must(r.SetSize([2]float64{300, 300}))
		must(r.SetRoundness(40))
		s, _ := l.RootGroup().AddStroke()
		must(s.SetColor(white))
		must(s.SetWidth(16))
		must(s.SetLineCap(aep.StrokeLineCapRound))
		must(s.SetLineJoin(aep.StrokeLineJoinRound))
	}
	// Position 关键帧动画
	{
		l := layer("⑥ Position 关键帧 AddKeyframeLinear", [2]float64{0, 0})
		r, _ := l.RootGroup().AddRect()
		must(r.SetSize([2]float64{120, 120}))
		f, _ := l.RootGroup().AddFill()
		must(f.SetColor(blue))
		must(l.Position().Clear())
		must(l.Position().AddKeyframeLinear(0, [2]float64{900, 560}))
		must(l.Position().AddKeyframeLinear(3, [2]float64{1720, 560}))
	}
	// Position 缓动动画
	{
		l := layer("⑦ Position 缓动 AddKeyframeWithEase(影响75%)", [2]float64{0, 0})
		r, _ := l.RootGroup().AddRect()
		must(r.SetSize([2]float64{120, 120}))
		f, _ := l.RootGroup().AddFill()
		must(f.SetColor(red))
		ease := aep.TemporalEase{Speed: 0, Influence: 75}
		must(l.Position().Clear())
		must(l.Position().AddKeyframeWithEase(0, [2]float64{900, 760}, ease, ease))
		must(l.Position().AddKeyframeWithEase(3, [2]float64{1720, 760}, ease, ease))
	}
	// Rotation/Scale/Opacity 关键帧
	{
		l := layer("⑧ Rotation/Scale/Opacity 关键帧", [2]float64{960, 980})
		r, _ := l.RootGroup().AddRect()
		must(r.SetSize([2]float64{200, 200}))
		f, _ := l.RootGroup().AddFill()
		must(f.SetColor(white))
		must(l.Rotation().AddKeyframeLinear(0, 0))
		must(l.Rotation().AddKeyframeLinear(4, 360))
		must(l.Scale().AddKeyframeLinear(0, [2]float64{50, 50}))
		must(l.Scale().AddKeyframeLinear(4, [2]float64{130, 130}))
		must(l.Opacity().AddKeyframeLinear(0, 30))
		must(l.Opacity().AddKeyframeLinear(4, 100))
	}

	f, err := os.Create(out)
	must(err)
	must(p.WriteAEP(f))
	must(f.Close())

	cwd, _ := os.Getwd()
	fmt.Printf("已生成: %s (cwd=%s)\n", out, cwd)
	fmt.Printf("合成 1 个，形状图层 %d 个\n", len(comp.Layers))
}
