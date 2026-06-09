// internal/aep/framerate_canonical.go
package codec

import "math"

// FrameRateEncoding 表示 cdta @0x9C..@0x9F 的 (Whole, Frac) 对编码。
// AE 存 frame rate 为 Whole + Frac/65536，但 NTSC fractions（23.976/29.97/59.94）
// 必须用 AE 内部 canonical 值，否则 AE 重新保存时 rewrite + 测试 flaky。
type FrameRateEncoding struct {
	Whole uint16
	Frac  uint16
}

// ntscCanonical: 容差 < ntscTolerance 内的用户 fps 都映射到这些 canonical 值。
// Frac 值实证来源:
//
//	29.97 / 59.94 来自 test_data/re_tickrate.aep RE_fps_29_97 / RE_fps_59_94 cdta @0x9E
//	23.976 数学计算 (24000/1001 - 23) * 65536 ≈ 63962 = 0xF9DA (无 fixture 实证但容差内安全)
const ntscTolerance = 1e-3

var ntscCanonical = map[float64]FrameRateEncoding{
	23.976: {Whole: 23, Frac: 0xF9DA},
	29.97:  {Whole: 29, Frac: 0xF852},
	59.94:  {Whole: 59, Frac: 0xF0A4},
}

// EncodeFrameRate 把用户输入 fps（如 29.97 或 30000/1001）normalize 为
// canonical (Whole, Frac) 二元组写入 cdta。
//
// 行为：
//   - 容差内匹配 ntscCanonical 表 → 用 canonical 值
//   - 否则走通用 round：Whole = floor(fps), Frac = round((fps - Whole) * 65536)
//
// 返回的 (Whole, Frac) 写入 cdta @cdtaFrameRateWhole / @cdtaFrameRateFrac。
func EncodeFrameRate(fps float64) FrameRateEncoding {
	for canonicalFps, enc := range ntscCanonical {
		if math.Abs(fps-canonicalFps) < ntscTolerance {
			return enc
		}
	}
	Whole := uint16(math.Floor(fps))
	Frac := uint16(math.Round((fps - float64(Whole)) * 65536))
	return FrameRateEncoding{Whole: Whole, Frac: Frac}
}

// DecodeFrameRate 反向解（给单测验证 round-trip）。
func DecodeFrameRate(enc FrameRateEncoding) float64 {
	return float64(enc.Whole) + float64(enc.Frac)/65536.0
}

// FpsTiming 是 cdta 里 fps-依赖的 secondary 字段集合。AE 25 写新 comp 时
// 这些字段必须填充正确值（实测 fixtures from re_tickrate.aep / re_cdta_probe.aep），
// 否则 AE 25 显示错误（duration/shutter）或拒绝打开。
//
// 实测两组 tick rate：
//   - TickRate (actual)  = TicksPerFrame × fps_actual (NTSC 含 .97/.94)
//   - NominalTickRate   = TicksPerFrame × fps_nominal_whole (NTSC nominal = 24/30/60)
//
// 整数 fps 下两者相等；NTSC 下不等。AE 用 NominalTickRate 当 @0x2C 的"duration in ticks"
// 单位（cdta @0x2C = duration_seconds × NominalTickRate），但 @0x08 / @0x18 / @0x30
// 存 actual TickRate（per RE_fps_29_97 / A_baseline 实证）。
//
// 字段 RE 来源:
//   - test_data/re_tickrate.aep 各 RE_fps_* comp 的 cdta @0x04..@0x33（actual TickRate）
//   - test_data/re_cdta_probe.aep A_baseline / F_shutter_angle_360（10s comps，验证
//     NominalTickRate × duration 公式）
type FpsTiming struct {
	TicksPerFrame   uint16 // cdta @0x06 (uint16 BE)
	TickRate        uint32 // cdta @0x08 / @0x18 / @0x30 (uint32 BE; = TicksPerFrame × fps_actual)
	NominalTickRate uint32 // cdta @0x2C divisor (uint32; = TicksPerFrame × fps_nominal_whole)
}

// canonicalFpsTiming: AE 25 实测 fps → TickRate / NominalTickRate 表。
//
// NTSC nominal_whole: 23.976→24 / 29.97→30 / 59.94→60（@0x2C 算 duration 用此 base，
// 不是 floor(fps)）。23.976 fixture 缺，用 1000 t/f 数学推算（其它 NTSC 实证）。
var canonicalFpsTiming = map[float64]FpsTiming{
	24:     {TicksPerFrame: 1024, TickRate: 24576, NominalTickRate: 24576},
	25:     {TicksPerFrame: 1024, TickRate: 25600, NominalTickRate: 25600},
	30:     {TicksPerFrame: 1024, TickRate: 30720, NominalTickRate: 30720},
	50:     {TicksPerFrame: 512, TickRate: 25600, NominalTickRate: 25600},
	60:     {TicksPerFrame: 512, TickRate: 30720, NominalTickRate: 30720},
	23.976: {TicksPerFrame: 1000, TickRate: 23976, NominalTickRate: 24000}, // no fixture; computed
	29.97:  {TicksPerFrame: 800, TickRate: 23976, NominalTickRate: 24000},
	59.94:  {TicksPerFrame: 400, TickRate: 23976, NominalTickRate: 24000},
}

// LookupFpsTiming 返回 fps 对应的 cdta timing 字段。
//
// 行为:
//   - canonical fps（容差 ntscTolerance 内）→ 表查找
//   - 其它 fps → 启发式:
//     TicksPerFrame = 1024 if fps ≤ 30 else 512
//     TickRate = round(TicksPerFrame × fps)
//     NominalTickRate = TicksPerFrame × round(fps) (假设非 NTSC)
//
// 启发式 fallback 对 AE 25 是否完全 safe 未实证 — 但比全零安全；
// 非 canonical fps 需进一步 RE 才能保证 AE 25 不 crash / 显示正确。
func LookupFpsTiming(fps float64) FpsTiming {
	for canonical, t := range canonicalFpsTiming {
		if math.Abs(fps-canonical) < ntscTolerance {
			return t
		}
	}
	var tpf uint16
	if fps <= 30 {
		tpf = 1024
	} else {
		tpf = 512
	}
	rounded := uint32(math.Round(fps))
	return FpsTiming{
		TicksPerFrame:   tpf,
		TickRate:        uint32(math.Round(float64(tpf) * fps)),
		NominalTickRate: uint32(tpf) * rounded,
	}
}
