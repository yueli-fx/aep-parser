// internal/aep/framerate_canonical.go
package aep

import "math"

// frameRateEncoding 表示 cdta @0x9C..@0x9F 的 (whole, frac) 对编码。
// AE 存 frame rate 为 whole + frac/65536，但 NTSC fractions（23.976/29.97/59.94）
// 必须用 AE 内部 canonical 值，否则 AE 重新保存时 rewrite + 测试 flaky。
type frameRateEncoding struct {
	whole uint16
	frac  uint16
}

// ntscCanonical: 容差 < ntscTolerance 内的用户 fps 都映射到这些 canonical 值。
// frac 值实证来源:
//   29.97 / 59.94 来自 test_data/re_tickrate.aep RE_fps_29_97 / RE_fps_59_94 cdta @0x9E
//   23.976 数学计算 (24000/1001 - 23) * 65536 ≈ 63962 = 0xF9DA (无 fixture 实证但容差内安全)
const ntscTolerance = 1e-3

var ntscCanonical = map[float64]frameRateEncoding{
	23.976: {whole: 23, frac: 0xF9DA},
	29.97:  {whole: 29, frac: 0xF852},
	59.94:  {whole: 59, frac: 0xF0A4},
}

// encodeFrameRate 把用户输入 fps（如 29.97 或 30000/1001）normalize 为
// canonical (whole, frac) 二元组写入 cdta。
//
// 行为：
//   - 容差内匹配 ntscCanonical 表 → 用 canonical 值
//   - 否则走通用 round：whole = floor(fps), frac = round((fps - whole) * 65536)
//
// 返回的 (whole, frac) 写入 cdta @cdtaFrameRateWhole / @cdtaFrameRateFrac。
func encodeFrameRate(fps float64) frameRateEncoding {
	for canonicalFps, enc := range ntscCanonical {
		if math.Abs(fps-canonicalFps) < ntscTolerance {
			return enc
		}
	}
	whole := uint16(math.Floor(fps))
	frac := uint16(math.Round((fps - float64(whole)) * 65536))
	return frameRateEncoding{whole: whole, frac: frac}
}

// decodeFrameRate 反向解（给单测验证 round-trip）。
func decodeFrameRate(enc frameRateEncoding) float64 {
	return float64(enc.whole) + float64(enc.frac)/65536.0
}

// fpsTiming 是 cdta 里 fps-依赖的 secondary 字段集合。AE 25 写新 comp 时
// 这些字段必须填充正确值（实测 fixtures from re_tickrate.aep / re_cdta_probe.aep），
// 否则 AE 25 显示错误（duration/shutter）或拒绝打开。
//
// 实测两组 tick rate：
//   - tickRate (actual)  = ticksPerFrame × fps_actual (NTSC 含 .97/.94)
//   - nominalTickRate   = ticksPerFrame × fps_nominal_whole (NTSC nominal = 24/30/60)
//
// 整数 fps 下两者相等；NTSC 下不等。AE 用 nominalTickRate 当 @0x2C 的"duration in ticks"
// 单位（cdta @0x2C = duration_seconds × nominalTickRate），但 @0x08 / @0x18 / @0x30
// 存 actual tickRate（per RE_fps_29_97 / A_baseline 实证）。
//
// 字段 RE 来源:
//   - test_data/re_tickrate.aep 各 RE_fps_* comp 的 cdta @0x04..@0x33（actual tickRate）
//   - test_data/re_cdta_probe.aep A_baseline / F_shutter_angle_360（10s comps，验证
//     nominalTickRate × duration 公式）
type fpsTiming struct {
	ticksPerFrame   uint16 // cdta @0x06 (uint16 BE)
	tickRate        uint32 // cdta @0x08 / @0x18 / @0x30 (uint32 BE; = ticksPerFrame × fps_actual)
	nominalTickRate uint32 // cdta @0x2C divisor (uint32; = ticksPerFrame × fps_nominal_whole)
}

// canonicalFpsTiming: AE 25 实测 fps → tickRate / nominalTickRate 表。
//
// NTSC nominal_whole: 23.976→24 / 29.97→30 / 59.94→60（@0x2C 算 duration 用此 base，
// 不是 floor(fps)）。23.976 fixture 缺，用 1000 t/f 数学推算（其它 NTSC 实证）。
var canonicalFpsTiming = map[float64]fpsTiming{
	24:     {ticksPerFrame: 1024, tickRate: 24576, nominalTickRate: 24576},
	25:     {ticksPerFrame: 1024, tickRate: 25600, nominalTickRate: 25600},
	30:     {ticksPerFrame: 1024, tickRate: 30720, nominalTickRate: 30720},
	50:     {ticksPerFrame: 512, tickRate: 25600, nominalTickRate: 25600},
	60:     {ticksPerFrame: 512, tickRate: 30720, nominalTickRate: 30720},
	23.976: {ticksPerFrame: 1000, tickRate: 23976, nominalTickRate: 24000}, // no fixture; computed
	29.97:  {ticksPerFrame: 800, tickRate: 23976, nominalTickRate: 24000},
	59.94:  {ticksPerFrame: 400, tickRate: 23976, nominalTickRate: 24000},
}

// lookupFpsTiming 返回 fps 对应的 cdta timing 字段。
//
// 行为:
//   - canonical fps（容差 ntscTolerance 内）→ 表查找
//   - 其它 fps → 启发式:
//     ticksPerFrame = 1024 if fps ≤ 30 else 512
//     tickRate = round(ticksPerFrame × fps)
//     nominalTickRate = ticksPerFrame × round(fps) (假设非 NTSC)
//
// 启发式 fallback 对 AE 25 是否完全 safe 未实证 — 但比全零安全；
// 非 canonical fps 需进一步 RE 才能保证 AE 25 不 crash / 显示正确。
func lookupFpsTiming(fps float64) fpsTiming {
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
	return fpsTiming{
		ticksPerFrame:   tpf,
		tickRate:        uint32(math.Round(float64(tpf) * fps)),
		nominalTickRate: uint32(tpf) * rounded,
	}
}
