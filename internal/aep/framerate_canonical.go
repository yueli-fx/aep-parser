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
