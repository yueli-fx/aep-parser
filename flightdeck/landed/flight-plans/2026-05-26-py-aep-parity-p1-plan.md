# py-aep parity — Phase 1 implementation plan

**Spec**: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md)
**Goal**: ship 50+ 新 API（全 reader / helper / single-byte writer），零 invariant 冲突。

**Out of scope for P1**: 任何结构性 chunk 增删（Layer.Remove/Duplicate/Move）、Render Queue、Essential Graphics、Gradient、CMS、nnhd 字节布局。

**完工标准**: PASS count +50 以上、`go vet ./...` clean、`docs/`/`coverage.md`/`coverage-detail.md`/`board.md` 同步、AE 2020+2025 ship gate 跑（针对 1D 写部分）。

---

## Task 1A — CompItem filter views（~15 API，纯 helper）

**目标**: 给 `Composition` 加按 layer 类型/source 类型过滤的 getter。

**实现位置**: `internal/aep/composition_views.go`（新文件）

**API surface**:
```go
func (c *Composition) TextLayers() []*Layer
func (c *Composition) ShapeLayers() []*Layer
func (c *Composition) CameraLayers() []*Layer
func (c *Composition) LightLayers() []*Layer
func (c *Composition) NullLayers() []*Layer
func (c *Composition) SolidLayers() []*Layer
func (c *Composition) AdjustmentLayers() []*Layer
func (c *Composition) ThreeDLayers() []*Layer
func (c *Composition) GuideLayers() []*Layer
func (c *Composition) SoloLayers() []*Layer
func (c *Composition) AVLayers() []*Layer
func (c *Composition) CompositionLayers() []*Layer  // source = Composition
func (c *Composition) FootageLayers() []*Layer       // source = Footage
func (c *Composition) FileLayers() []*Layer          // source = file
func (c *Composition) PlaceholderLayers() []*Layer   // source = placeholder
```

**测试**: `composition_views_test.go` — 跑 `re_batch*.aep`（已有混合 layer 类型），断言每个 filter 返回正确 subset。

**步骤**:
1. 实现 15 个 method（每个一行 filter）
2. 写 test
3. 同步 [docs/composition.md](../../docs/composition.md)
4. 更新 [coverage.md](coverage.md) Composition 段

**估算**: 半天

---

## Task 1B — Project filter views + LayerByID + EffectNames（~5 API）

**实现位置**: `internal/aep/project_views.go`（新文件）

**API**:
```go
func (p *Project) Folders() []*Folder
func (p *Project) Footages() []*Footage
func (p *Project) RootFolder() *Folder
func (p *Project) LayerByID(id uint32) *Layer
func (p *Project) EffectNames() []string   // from Pefl/pjef
```

**RE 工作**: `Pefl/pjef` chunk family — root-level LIST 含 `pjef` Utf8 chunks per effect name. 写 RE fixture 不需要（已 in re_batch.aep）。

**测试**: 校验 `re_batch.aep` `EffectNames()` 返回所有 effect match-name + `LayerByID` 跨 comp 查找正确。

**步骤**:
1. Pefl/pjef parse helper
2. 5 个 method
3. test
4. docs + coverage

**估算**: 1 天

---

## Task 1C — Frame-time 伴生 accessor（~20 API，纯 wrapper）

**实现位置**: 新文件 `internal/aep/frame_time_accessors.go` + 各 type 加 method

**API**（每个原 `XxxTime float64` 都加 `FrameXxxTime int`）:
```go
// Layer
func (l *Layer) FrameInPoint() int
func (l *Layer) SetFrameInPoint(frame int) error
func (l *Layer) FrameOutPoint() int
func (l *Layer) SetFrameOutPoint(frame int) error
func (l *Layer) FrameStartTime() int
func (l *Layer) SetFrameStartTime(frame int) error
func (l *Layer) FrameTime() int   // current time

// Composition
func (c *Composition) DisplayStartFrame() int
func (c *Composition) SetDisplayStartFrame(frame int) error
func (c *Composition) WorkAreaStartFrame() int
func (c *Composition) SetWorkAreaStartFrame(frame int) error
func (c *Composition) WorkAreaDurationFrame() int
func (c *Composition) SetWorkAreaDurationFrame(frame int) error
func (c *Composition) FrameDuration() int   // total dur in frames
func (c *Composition) FrameTime() int

// Keyframe
func (k *Keyframe) FrameTime() int
func (k *Keyframe) SetFrameTime(frame int) error

// Marker
func (m *Marker) FrameTime() int
func (m *Marker) SetFrameTime(frame int) error
func (m *Marker) FrameDuration() int
func (m *Marker) SetFrameDuration(frame int) error
```

**实现要点**: 用每个 type 所属 comp 的 `TickRate` 换算 — 已经在 parse_composition.go 提取。Layer 用 `l.containingComp().TickRate()`。Keyframe / Marker 用各自 owning property/comp。

**测试**: 跑现有 fixture，对照 seconds 字段做 round-trip ((s × tickRate / aeLegacyTimeBase × fps) → int 帧)。

**步骤**:
1. helper `secondsToFrames(s float64, tickRate uint32, fps float64) int` + `framesToSeconds(...)` 反向
2. 20 个 method
3. test
4. docs

**估算**: 1 天

---

## Task 1D — Project single-field chunks（~8 API，部分 R/W）

**实现位置**: `internal/aep/parse_project.go` + `internal/aep/write_project.go` (二者已存) + 加 chunk parsers

**新 chunk parse**:
- `head` chunk: revision field offset (RE 需要 — 找一个 fixture，AE 改一次 save 一次 diff revision byte)
- `lnrb` / `lnrp` flag chunks (root-level, presence = bool)
- `acer` (1 byte bool)
- `adfr` (8 byte f64)
- `dwga` (8 byte f64)
- `gpug` LIST → Utf8 (length-variable splice)
- `ExEn` LIST → Utf8 (length-variable splice)

**API**:
```go
func (p *Project) Revision() uint32                   // R only
func (p *Project) LinearBlending() bool
func (p *Project) SetLinearBlending(v bool)           // add/remove lnrb chunk
func (p *Project) LinearizeWorkingSpace() bool
func (p *Project) SetLinearizeWorkingSpace(v bool)
func (p *Project) CompensateForSceneReferredProfiles() bool
func (p *Project) SetCompensateForSceneReferredProfiles(v bool) error
func (p *Project) AudioSampleRate() float64
func (p *Project) SetAudioSampleRate(rate float64) error  // {22050,32000,44100,48000,96000}
func (p *Project) WorkingGamma() float64
func (p *Project) SetWorkingGamma(g float64) error    // {2.2, 2.4}
func (p *Project) GpuAccelType() GpuAccelType        // enum
func (p *Project) SetGpuAccelType(t GpuAccelType) error
func (p *Project) ExpressionEngine() string
func (p *Project) SetExpressionEngine(eng string) error  // {"extendscript","javascript-1.0"}
```

**RE 工作量**: 
- `revision`: 单 fixture 跑 AE 改 N 次 save，head chunk byte diff 找 4B BE counter 位置
- `acer/adfr/dwga`: 写一个 jsx 把 project setting 拉 max/min，diff chunk bytes
- 其它通过 chunk presence/text 判定

**ship gate**: `SetLinearBlending(true)` + `SetExpressionEngine("javascript-1.0")` 后用 AE 2020/2025 双开校验。

**步骤**:
1. 写 `re_v2_3_project_chunks.jsx` (拉 8 个 setting 全 max/min，存 fixture)
2. dump tool: `tmp_debug/dump_project_settings/main.go`
3. RE 各 chunk byte location
4. 实现 R/W
5. ship gate 验
6. docs + coverage + scars (有 negative finding 时)

**估算**: 1-2 天（RE 1 天 + 实现 1 天）

---

## Task 1E — Layer convenience（~15 API，pure helper）

**实现位置**: `internal/aep/layer_convenience.go`

**API**:
```go
func (l *Layer) Index() int                  // 0-based within comp.Layers
func (l *Layer) LayerType() string           // "AVLayer" / "CameraLayer" / "LightLayer" / "TextLayer" / "ShapeLayer"
func (l *Layer) HasVideo() bool
func (l *Layer) HasAudio() bool
func (l *Layer) AudioActive() bool
func (l *Layer) Active() bool                // visible && enabled && in time range
func (l *Layer) ActiveAtTime(t float64) bool
func (l *Layer) AudioActiveAtTime(t float64) bool
func (l *Layer) Width() int                  // proxy to source
func (l *Layer) Height() int
func (l *Layer) HasTrackMatte() bool         // 自己有 matte
func (l *Layer) IsTrackMatte() bool          // 是别人的 matte
func (l *Layer) AutoName() string            // source.Name or other auto
func (l *Layer) IsNameFromSource() bool
func (l *Layer) ContainingComp() *Composition
func (l *Layer) RemoveTrackMatte() error     // splice trackMatte fields to 0
```

**测试**: re_batch*.aep 上跑 — 每个 helper 跟 direct field 计算结果 match.

**估算**: 0.5 天

---

## Task 1F — Composition convenience（~5 API）

**实现位置**: `internal/aep/composition_convenience.go`

**API**:
```go
func (c *Composition) NumLayers() int           // len(c.Layers)
func (c *Composition) HasAudio() bool           // any layer with audio_enabled
func (c *Composition) ActiveCamera() *Layer     // first camera with active=true
func (c *Composition) Markers() []*Marker       // flat list, alias of comp marker keyframes
func (c *Composition) TimeScale() float64       // alias of TickRate
```

**估算**: 0.5 天

---

## Task 1G — Property tdb4 flag readers（~7 API）

**实现位置**: `internal/aep/parse_properties.go` 加 method on `*Property`

**API**:
```go
func (p *Property) IsSpatial() bool      // tdb4 is_spatial bit
func (p *Property) IsAnimated() bool     // len(p.Keyframes) > 0
func (p *Property) IsColor() bool        // Components == 4 + tdb4 color bit
func (p *Property) IsInteger() bool      // tdb4 integer bit
func (p *Property) IsNoValue() bool      // tdb4 no_value bit
func (p *Property) IsVector() bool       // tdb4 vector bit
func (p *Property) CanVaryOverTime() bool  // !is_no_value && !is_runtime_only
```

**RE 工作**: tdb4 字节布局 — `is_spatial / color / integer / no_value / vector` 各 1 bit。py-aep 的 `binary/property_chunks.py` 有 mapping，参考即可。

**测试**: re_batch*.aep 上 — Tritone 颜色 IsColor==true、IsSpatial==false；Position IsSpatial==true。

**估算**: 0.5 天

---

## Task 1H — Footage convenience（~5 API）

**实现位置**: `internal/aep/footage_convenience.go`

**API**:
```go
func (f *Footage) FootageMissing() bool
func (f *Footage) HasAudio() bool
func (f *Footage) StartFrame() int       // RE'd from sspc?
func (f *Footage) EndFrame() int
func (f *Footage) AssetType() string     // "placeholder" / "solid" / "file"
func (f *Footage) File() string           // convenience, same as Path
```

**RE 工作**: StartFrame/EndFrame 在 sspc 字节布局未 RE 的部分 — 看 py-aep `parsers/footage.py`。

**估算**: 0.5-1 天

---

## Task 1I — Application wrapper + Version（~3 API）

**实现位置**: `internal/aep/application.go`（新文件）

**API**:
```go
type Application struct {
    Project *Project
    // ... runtime fields stay nil/unsupported
}

func Parse(path string) (*Application, error)    // alias of Open + wrap
func ParseReader(r io.Reader) (*Application, error)

func (a *Application) Version() string           // from head chunk (RE'd in V2.1)
```

**目标**: API 表面跟 py-aep `app = py_aep.parse("x.aep")` 对齐，**不破坏 `aep.Open()` 现有 entry**。

**测试**: `parse_application_test.go` — version 字段验证 + Application.Project 等价于 Open() 返回。

**估算**: 0.5 天

---

## P1 总估算

| Task | 估算 | API 数 |
|---|---|---|
| 1A CompItem filter | 0.5 d | 15 |
| 1B Project filter + LayerByID | 1 d | 5 |
| 1C Frame-time accessor | 1 d | 20 |
| 1D Project single-field chunks | 1.5 d | 8 (R+W) |
| 1E Layer convenience | 0.5 d | 15 |
| 1F Composition convenience | 0.5 d | 5 |
| 1G Property tdb4 flag | 0.5 d | 7 |
| 1H Footage convenience | 0.5-1 d | 5 |
| 1I Application wrapper | 0.5 d | 3 |
| **总** | **5-7 d** | **~83 API** |

每 task 单 commit：`feat(<scope>): py-aep P1 <task-id> — <hook>`

---

## P1 完工 checklist

- [ ] 9 个 Task 全 ship
- [ ] PASS count +50 以上
- [ ] `go vet ./...` clean
- [ ] `docs/`：composition / layer / project / footage / property 全同步
- [ ] [coverage.md](coverage.md) Project / Composition / Layer / Footage / Property 段 update
- [ ] [coverage-detail.md](coverage-detail.md) AE attr 表 update
- [ ] [board.md](../board.md): "最近归档" + Active focus 滚到 P2
- [ ] Task 1D ship gate: AE 2020 + AE 2025 双开 `SetLinearBlending/SetExpressionEngine` 后 file 校验通过
- [ ] Phase 1 done → 起 `2026-05-XX-py-aep-parity-p2-plan.md`

---

## 相关
- Spec: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md)
- 参照源: [`../reference/py-aep/`](../reference/py-aep/)
