# Coverage 概览（精简入口）

> 要查具体 AE attribute → Go 字段对应，看 [coverage-detail.md](coverage-detail.md) 详细交叉表。
> 本文档只列：什么已 ship / 什么暂搁 / 什么不可达 / negative findings。

兼容声明：AE 2020 读下限 + AE 24+ 字段渐进写。所有 setter 都是 length-preserving（除少数 length-variable 替换：name / comment / expression / 字体名 / 文本内容）。

测试基线 / PASS count 见 `../board.md` `Last updated` 行（**唯一权威**）。

---

## ✅ 已 ship

### V2 结构性创建 (2026-05-22)

- `aep.NewProject(target ...AETarget) *Project` — 全新空 project，零参 = TargetAE2020；支持 TargetAE2020 / 2022 / 2025
- `proj.NewComposition(name, w, h, fps, duration) (*Composition, error)` — 在 root folder 新建空 comp，原子（warning/error → rollback），跟 `Open(...)` 出来的 comp 同构（所有 Set\* 立即可用）
- AE 2020 + AE 2025 ship gate PASS（`AE_SHIP_GATE=1 go test -run TestV2_1_AEShipGate -v`）
- 详 `../scars/ae25-acceptance-gate.md` 5 阶段 RE findings + canonical seed 策略

### Project / Composition / Item

- Project: `BitsPerChannel` R/W
- Composition（cdta）: `Name / FrameRate / Duration / Size / BGColor / ShutterAngle / ShutterPhase / MotionBlurAdaptive / MotionBlurSamplesPerFrame / WorkArea / DisplayStartTime / DisplayStartFrame / PixelAspect / ResolutionFactor` R/W
- Composition（PRin LIST）: `Renderer` **R only**（match-name；`ADBE Escher` = Advanced 3D / `ADBE Ernst` = Cinema 4D / `ADBE Standard` = Classic 3D）
- Composition（cdta flag bits）: `HideShyLayers / CompMotionBlur / Draft3D / FrameBlending / PreserveNestedFrameRate / PreserveNestedResolution` R/W
- Item: `Name / Comment / Label` R/W（Composition + Footage 共用）
- Footage: `Path` R/W
- Composition markers: 全 8 个 setter
- **Composition filter views (py-aep parity P1 1A)**: `TextLayers / ShapeLayers / CameraLayers / LightLayers / NullLayers / AdjustmentLayers / ThreeDLayers / GuideLayers / SoloLayers / AVLayers / CompositionLayers / FootageLayers / FileLayers / SolidLayers / PlaceholderLayers` — 15 个 filter helper，无写
- **Composition convenience (P1 1F)**: `NumLayers / HasAudio / TimeScale` — 3 个 helper（ActiveCamera + Markers field 早已 ship）
- **Footage discriminator**: 新增 `IsPlaceholder` 字段（opti tag = "Plac"）;`Footage.{IsSolid, IsPlaceholder}` 互斥三态（file / solid / placeholder）

### Layer

- 基础（ldta）: `Name / Comment / Label / InPoint / OutPoint / StartTime / Stretch / Parent / Source / TrackMatte / TrackMatteLayer / AutoOrient` R/W
- Flag bit: `Visible / Solo / Shy / Locked / EffectsEnabled / MotionBlur / AudioEnabled / FrameBlend{Enabled,PixelMotion} / CollapseTransform / Is3D / IsAdjust / IsGuide / IsNull / MarkersLocked / PreserveTransparency / Quality / SamplingBicubic / BlendingMode` R/W
- Markers: 全 8 个 setter
- AlternateSource（AE 18+ EGP slot）: R/W；首次启用 = 结构性 refused
- **Transform**: 8 typed setter (`SetAnchorPoint / SetPosition / SetScale / SetRotation / SetRotateX / SetRotateY / SetOrientation / SetOpacity`)
- **AudioLevels**: `SetAudioLevels([L, R])`
- **Camera 专属**: 13 typed accessor pair (`CameraZoom / DepthOfField / FocusDistance / Aperture / BlurLevel + Iris × 8`)，`SetCameraDepthOfField(bool)` 自动 0/1
- **Light 专属**: 11 typed accessor pair (`LightColor / Intensity / ConeAngle / ConeFeather / FalloffType / FalloffStart / FalloffDistance / CastsShadows / ShadowDarkness / ShadowDiffusion`) + `LightKind` ldta `@0x88` R/W；`SetLightCastsShadows(bool)`
- **Material Options 3D AV**: 17 typed accessor pair + `MaterialCastsShadowsMode` 三态 enum（Off/On/Only）
- **Geometry Options 3D AV**: 3 typed accessor pair（PlaneCurvature / PlaneSubdivision / BevelDirection）

### Property / Keyframe

- `StaticValue` R/W（对 effect 参数也直接生效）/ `Expression` 完整 R/W（length-variable）/ `ExpressionEnabled` R/W（反语义解析）
- Keyframe: `Time / Value / InInterp / OutInterp / Temporal{In,Out}Ease / Spatial{In,Out}Tangent` 全 R/W（对 effect 参数 keyframe 也直接生效）
- 增删 keyframe: `InsertKeyframe(time, value)` + `DeleteKeyframe(i)`（要求 ≥ 1 既有 keyframe 作 layout 模板）

### Mask

- 顶层: `Mode / Inverted / Color / Closed / Locked / MotionBlur` R/W；`Feather / Opacity / Expansion` R + 顶层 setter
- 路径动画: `PathKeyframes` R only（顶点重写 = 结构性 ❌）

### Shape

- 自由路径: `Layer.ShapePaths` R only
- 参数化 Rect / Ellipse / Star: `ShapePrimitives` R/W（子字段 `*Property`）

### V2.2 alpha ShapeLayer 写路径 (2026-05-25, iter-7/8)

- `(c *Composition) NewShapeLayer(name string) (*ShapeLayer, error)` — 新建空 ShapeLayer，AE 2025 接受 + comp.layers.length=1
- `(s *ShapeLayer) RootGroup() *VectorGroup` — 取顶层 Contents 容器
- `(g *VectorGroup) AddRect() (*RectNode, error)` / `AddFill() (*FillNode, error)` — 加入参数化 shape kid
- Setter: `RectNode.SetSize([w, h]) / FillNode.SetColor([r,g,b,a])` static-only
- **AE 接受 gate**: 通过 embed tolerance.aep 抽出的 3 处 boilerplate 字节 (Transform Group 1842B + Rect body 448B + Fill body 426B in `internal/aep/templates/`)；详 `../scars/v2-2-aelayer-structure.md`
- **V2.2 alpha 限制**（V2.2.1 候选）:
  - Ellipse / Path / Stroke: Go 端能 emit + parse，AE 会 silent drop（需各自 fixture + embed bytes）
  - Fill Color 编码: cdat scalar 跟 JSX 0-1 input 不对齐（tolerance 0.5 → 0x406fe0... ≈ 255），可见色可能错
  - Keyframe 持久化（Rect Size / Fill Color / Layr Position 全部）: 不持久化，first kf 作 static fallback
  - Layr Transform 的 Anchor / Scale / Rotation / Opacity: runtime-only 不持久化
  - Rect Direction / Position / Roundness: runtime-only 不持久化

### Text

- 基础: `Text` R/W (length-preserving)；`Fonts []string` + `AddFont(name)` 追加；`FontIndex` R/W
- per-run（22 setter）: `FontSize / FillColor / StrokeColor / StrokeWidth / ApplyStroke / Tracking / Leading / AutoLeading / BaselineShift / HorizontalScale / VerticalScale / Tsume / FauxBold / FauxItalic / CapsOption / BaselineOption / StrokeOverFill / AutoKernType / NoBreak / LineJoinType / DigitSet`
- per-paragraph（9 setter）: `Justification / FirstLineIndent / StartIndent / EndIndent / SpaceBefore / SpaceAfter / AutoHyphenate / LeadingType / HangingRoman / Direction`
- btdk-root: `SetManualKerning(values []int)`（要求 `/8` slot 已 emit）
- 字体 axes: `FontAxes [][]float64` R only（写绑定到 PostScript 名切换）

---

## 🗑️ 暂搁（fixture / 环境阻塞，遇到需要再做）

| 项 | 阻塞原因 |
| --- | --- |
| `AVLayer.environmentLayer` | 需 equirectangular 360° 视频素材 |
| `TextDocument.ligature` | 默认字体 ligature=false 设值无 diff；需带 OT `liga` feature 的字体 fixture |
| `maskFeatherFalloff` | JSX 设值后 mkif 字节零变化；疑似在 mask sub-property 树，需深挖 RE |

## ❌ 不可达（length-preserving 写约束之外 / AE 限制）

| 项 | 原因 |
| --- | --- |
| Layer / Effect / Mask vertex / ShapePrimitive **增删** | 结构性，破坏多个父 LIST 大小 |
| `Property.timeRemapEnabled` toggle | AE 加/删 2 个 identity keyframe（结构性） |
| `Property.dimensionsSeparated` | AE 拆 Position 成 3 个 1D 属性（+306 字节非局部改动） |
| `lineOrientation` 横/竖排切换 | layer-local 坐标重排 + 多字段连锁 |
| `MaskPropertyGroup.rotoBezier` | 切换重写整个 shape 顶点表示（+16 字节，4500+ byte-diff） |
| Motion Graphics Template / EP 模板 binding（除 `alternateSource`） | 跨 chunk 复杂结构，P3 罕用 |
| Project 渲染设置（`gpuAccel / colorSpace / expressionEngine`） | P3，AE 24+ 大多锁定为 default。`renderer` 已 ship R；setter 仍是 P3（prin 双段 NUL-sep + prda 长度随 renderer 变） |
| Adobe World-Ready composer 切换 | P3 |
| 手动 kerning **首次启用** | 结构性添加（需 AE 先 emit `/8` slot） |
| Camera `FilmSize` setter | ldta `@0x98` 持久化但 ScriptingAPI 不暴露写路径（详 `scars/camera-filmsize-ldta-write-blocked.md`） |

## ⚠ Negative findings（runtime-only / AE ScriptingAPI 限制）

| 项 | 结论 |
| --- | --- |
| `CompItem.dropFrame` | AE 不持久化（脚本可设可读，字节零变化）—— FrameRate NTSC 自动推断 |
| `TextDocument.fontLocation` | AE 不持久化（runtime 从系统字体注册表 join 路径） |
| Variable fonts axes **写** | `TextDocument.fontVariation` 不存在；唯一写途径 = 切已加载的 named-instance |
| `composerEngine` / `everyLineComposer` | AE 24+ 只接受 UNIVERSAL；其它写入抛错 |
| `setAlternateSource(item)` AE 脚本行为 | 自动包 wrapper precomp；我们 setter 不包装 |
| AE 24/25 ldta 加长 | 实测仍 164 字节（与 AE 23+ 一致），未见第三种长度 |
| cdta tail (≥ 0xCC) | 不存在 —— cdta 总长就是 0xCC=204 |

---

## ❓ 剩余可探方向（非 candidate 列表，需要新发现才动）

可达字段约 99% 已 ship。继续动需要：

1. **同-renderer 范围内 `Composition.SetRenderer`** — 跨 renderer 切换属结构性（prda 长度变），同 renderer 安全；价值小。
2. **ldta `@0x60-0x82` / `@0x8C-0x9F` 零值区 probe** — 高密度 JSX layer-flag 探针，可能挖出 1-2 个零散 flag 或全 negative。
3. **Footage proxy 字段** — 大部分结构性。
4. **Project nhed/nnhd 扩展字段** — 除 BitsPerChannel 外的字节，可能持 ColorSpace / Working Color Profile。
