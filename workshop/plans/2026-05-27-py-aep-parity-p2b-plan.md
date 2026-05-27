# py-aep parity — Phase 2b implementation plan

**Spec**: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md) § 2.1-2.5
**Predecessor**: [P2a plan](2026-05-27-py-aep-parity-p2a-plan.md) (completed 2026-05-27)
**Goal**: ship P2 中等域 — nnhd byte layout RE + CMS JSON + Gradient XML + Property metadata

**Out of scope for P2b**: PropertyGroup 链式访问 (P2c)、TimeRemap enable (P2d, 等 V3 capability framework)

**完工标准**: PASS count +10 以上、`go vet ./...` clean、`docs/{project,property}.md` + `coverage.md` + `coverage-detail.md` + `board.md` 同步

**Pre-flight check**:
```powershell
go vet ./...                                                      # clean
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS' # 213 (P2a baseline)
git status --short                                                # clean
```

---

## Task 1 — 2A nnhd byte layout RE (8 fields)

**目标**: 解析并暴露 nnhd chunk 中的 8 个字段。

**实现位置**: `internal/aep/project_settings.go`（append）

**API surface**:
```go
// FeetFramesFilmType: MM16 (16mm) or MM35 (35mm)
func (p *Project) FeetFramesFilmType() FeetFramesFilmType
func (p *Project) SetFeetFramesFilmType(v FeetFramesFilmType) error

// FootageTimecodeDisplayStartType
func (p *Project) FootageTimecodeDisplayStartType() FootageTimecodeDisplayStartType
func (p *Project) SetFootageTimecodeDisplayStartType(v FootageTimecodeDisplayStartType) error

// TimecodeDefaultBase: 1-999
func (p *Project) TimecodeDefaultBase() int
func (p *Project) SetTimecodeDefaultBase(v int) error

// FramesCountType
func (p *Project) FramesCountType() FramesCountType
func (p *Project) SetFramesCountType(v FramesCountType) error

// DisplayStartFrame: 0 or 1 (derived from frames_count_type % 2)
func (p *Project) DisplayStartFrame() int
func (p *Project) SetDisplayStartFrame(v int) error

// FramesUseFeetFrames: bool
func (p *Project) FramesUseFeetFrames() bool
func (p *Project) SetFramesUseFeetFrames(v bool) error

// TimeDisplayType
func (p *Project) TimeDisplayType() TimeDisplayType
func (p *Project) SetTimeDisplayType(v TimeDisplayType) error

// TransparencyGridThumbnails: bool
func (p *Project) TransparencyGridThumbnails() bool
func (p *Project) SetTransparencyGridThumbnails(v bool) error
```

**RE prereq**: py-aep `item_chunks.py:252-300` NnhdChunk 已完整定义字节布局。

**nnhd byte layout** (40 bytes):
- Bytes 0-7: reserved
- Byte 8: `_display_byte`
  - bit 7 = feet_frames_film_type (0=MM35, 1=MM16)
  - bits 6-0 = time_display_type (0=TIMECODE, 1=FRAMES)
- Byte 9: footage_timecode_display_start_type (0=FTCS_START_0, 1=FTCS_USE_SOURCE_MEDIA)
- Byte 10: reserved
- Byte 11: `_feet_byte`
  - bit 0 = frames_use_feet_frames
- Bytes 12-13: reserved
- Bytes 14-15: timecode_default_base (u2 BE)
- Bytes 16-19: unknown (default 0x00000010)
- Byte 20: frames_count_type (0=FC_START_0, 1=FC_START_1, 2=FC_TIMECODE_CONVERSION)
- Bytes 21-23: reserved
- Byte 24: bits_per_channel (already implemented)
- Byte 25: transparency_grid_thumbnails (bool)
- Bytes 26-39: unknown

**步骤**:
1. Add enum types to `internal/aep/enums.go`:
   - `FeetFramesFilmType` (MM16=2412, MM35=2413)
   - `FootageTimecodeDisplayStartType` (FTCS_USE_SOURCE_MEDIA=2212, FTCS_START_0=2213)
   - `FramesCountType` (FC_START_0=2612, FC_START_1=2613, FC_TIMECODE_CONVERSION=2614)
   - `TimeDisplayType` (TIMECODE=2012, FRAMES=2013)
2. Add getters/setters to `internal/aep/project_settings.go`
3. Write roundtrip tests
4. Sync docs + coverage

**估算**: 2 hr

---

## Task 2 — 2B CMS JSON (AE 24+)

**目标**: 解析并暴露 CMS (Color Management System) 相关字段。

**API surface**:
```go
func (p *Project) ColorManagementSystem() string
func (p *Project) SetColorManagementSystem(v string) error

func (p *Project) LutInterpolationMethod() string
func (p *Project) SetLutInterpolationMethod(v string) error

func (p *Project) OcioConfigurationFile() string
func (p *Project) SetOcioConfigurationFile(v string) error

func (p *Project) WorkingSpace() string  // R only
func (p *Project) DisplayColorSpace() string  // R only
```

**RE prereq**: 需要 AE 24+ fixture with CMS settings.

**估算**: 1-2 hr (depends on fixture availability)

---

## Task 3 — 2C Gradient XML

**目标**: 解析 ADBE Vector Grad Colors cdat 中的 XML。

**API surface**:
```go
type Gradient struct {
    ColorStops []ColorStop
    AlphaStops []AlphaStop
}

func (p *Property) Gradient() (*Gradient, error)
func (p *Property) SetGradient(g *Gradient) error
```

**RE prereq**: 需要 fixture with gradient effect.

**估算**: 2-3 hr

---

## Task 4 — 2D Property metadata (remaining)

**目标**: 暴露 Property 元数据字段。

**API surface**:
```go
func (p *Property) MinValue() float64
func (p *Property) MaxValue() float64
func (p *Property) UnitsText() string
func (p *Property) DefaultValue() float64
func (p *Property) LastValue() float64
func (p *Property) NbOptions() int
func (p *Property) PropertyControlType() int
func (p *Property) PropertyValueType() int
```

**RE prereq**: 需要 pard chunk reader + specs.py schema table.

**估算**: 2-3 hr

---

## 执行顺序

1 → 2 → 3 → 4。Task 1 是 nnhd RE，最基础，先做。

每个 task 完工 commit `feat(<scope>): py-aep parity P2b Task N — <desc>`；最后 docs/board 闭环 commit `docs(workshop): py-aep parity P2b 全闭环 / archive plan`。

---

## Self-review checklist

| Spec section | Plan coverage |
|---|---|
| 2.1 Project — nnhd fields (8) | Task 1 |
| 2.1 Project — CMS JSON | Task 2 |
| 2.4 Property — Gradient XML | Task 3 |
| 2.4 Property — metadata (MinValue/MaxValue/etc) | Task 4 |

**Deferred to P2c / 后续**：
- 2G PropertyGroup 链式访问（推 P2c）
- 2K TimeRemap enable（推 P2d，等 V3 capability framework）
