---
status: active
when_to_read: 判断某 chunk/字段属于哪个 AE 版本的版本门禁；想用 head 字节判 AE 版本；调查 ldta 160/164 尾长之谜；评估或参考外部 AE 降级工具的可信度与已知 bug；需要 AE ship-gate 自动化的 prefs/occlusion 注意事项
applies_to: [version-gating, ppSn, pdvc, head-fingerprint, ldta-length, external-tool-re, ship-gate, reference]
last_updated: 2026-06-09
resolved_by:
---

# 外部 AE 版本降级器 CEP 工具 RE

## Signature
- symptom: `某 chunk/字段是哪个 AE 版本引入的？head 怎么判版本？ldta 为何有 160/164 两种长度？外部降级器可信吗？`
- error_type: —（reference / 外部工具逆向，非本仓 bug）
- where: 外部工具 `ae-version-downgrader-cep/js/converter.js`；本仓 version-gating + ldta 知识
- trigger: 需要版本门禁判断 / 参考外部 AE 降级工具 / 调查 ldta 尾长

## 这是什么

逆向了一个**外部付费 CEP 扩展** `ae-version-downgrader-cep`（把高版本 `.aep` 降级成低版本可打开）。不是本仓代码，但它的版本门禁判断对 aep-parser 有交叉佐证价值。完整逐函数细节在个人 memory `reference_ae_version_downgrader_cep.md`（含反混淆复现法、全字节表）；本 incident 只留**对 aep-parser 有用的已验证结论**。三层验证全过：静态反混淆 `converter.js` → node `vm` 里 live 跑真 `convertFile`（8 例 + 2 throw，副本操作）→ AE 真接受 ship-gate。

## 对 aep-parser 有用的已验证发现

1. **版本指纹 = `head` chunk data 的 byte[1]/[3]/[4]**（独立字节 walker 扫 180+ 真实存盘验证）：`5c/_/07`=2018、`5d/16/0b`=2020、`5d/1d/0b`=2021、`5d/2b/0b`=2022、`5e/04/0b`=2023(真实)、`5f/_/0f`=2024、`60/01/0f`=2025；**兜底 `head[1]-0x47=major`**（detect 即便精确行不中也能靠这个兜回版本）。注意 AE 内部 major 跳号：2021=18(0x12) 直接跳到 2022=22(0x16)。
2. **`ppSn` 与 `pdvc` 都是 AE2022 引入的版本门禁 chunk**：≤2021 文件零出现，2022+ 才有。`pdvc` 每工程 1 个；`ppSn` 内容相关（计数随 comp/属性增长，1comp=1 / 2comp=2，实测）。`mrid`/`head`/`nhed`/`nnhd`/`svap` 是所有版本通用 chunk。
3. **`ldta` 有 160/164 两种长度，边界在 AE2023**（解开 [ldta-length-third-variant-not-found.md](ldta-length-third-variant-not-found.md) 的 negative finding 的另一半）：
   - **原生 AE2020 = ldta 160 字节**（本仓 fixture 实测：`2020_dummy_comp` 160×11、`tmp_debug/ellipse_ae2020_native` 160×12、`renderer_ae2020_r0/r1` 160×12）。
   - **AE2025 = ldta 164 字节**（`selection_both_layers` 实测 164×13）。
   - 即 AE2023+ 在 ldta 尾部加了 **4 字节**（落在 `@0xA0`=160..163，正是该 incident 列的 Wave-2 字段位）；≤2022 没有这 4 字节。降级器对 source≥2023→target≤2022 正是**砍掉这尾 4 字节**（164→160），ship-gate 实测 AE2020 无损接受 → 反证 160 就是 ≤2022 原生长度。**精确 pivot=2023 是工具断言**（本仓 2021–2024 模板都是空工程、0 ldta，未独立量到原生分层 2021/2022/2023 文件；已量到的是 2020=160、2025=164）。

## 该工具已证实的 5 个 bug（即"它有问题吗"——有，但都在边缘 target/输入）

1. `ppSn` 护栏（target≤2021 且 source≥2022 且没删到 ppSn 就 throw）**误杀无 ppSn 的精简工程/空模板**（`AE2025_empty`、空 2022 模板实测 throw）。
2. 降到 **2021 残留 `pdvc`**：pdvc 删除门 target≤2020，比 ppSn 门 target≤2021 低一档 → 2022/2025→2021 留下 pdvc，真实 2021 无此 chunk（实测 1comp→2021 后 pdvc 仍=1）。
3. 写 2023 时 `head[3]` 写成 **0x09**，真实 2023 存盘是 **0x04**（cosmetic；detect 有 `head[1]-0x47` 兜底不受影响，但写出的字节与真 2023 不符）。
4. 2018/2019 档删 `mrid`/`idpc`/`iide`/`comr` **不可信**：这些 chunk 在所有真实文件（含 2020）都存在（实测 AE2025 工程 `comr`/`CIF3` 确在），作者自标 experimental/beta。
5. 2022 目标的 `trimmedLdta` 护栏（找不到 164-ldta 可砍就 throw）**误杀无 shape/layer-ldta 的工程**。

## 验证方法（三层互证，无矛盾）

- 静态：node `vm` stub `require`+补 `Buffer`，只跑纯查表解码器反混淆（不调 convertFile），逐函数读出偏移/阈值/护栏。
- live：同 vm 真 fs/path 调真 `convertFile` 对 fixture 副本转换，自写字节 walker diff —— 朴素偏移 `head.dataStart+[1,3,4,5,6,7]`+write-if-different、trimmedLdta 164→160、comr/CIF3/OvdG 删增、bug#1/#2 全部活证。
- ship-gate：`selection_both_layers`(AE2025, 2 图层)→2020 经 `scripts/ae_run.ps1` 开 AE 2020 → 4 items / Comp 1 / 2 层同名同序，无"损坏/跳过"，**工具主流路径(2025→2020)产出 AE 视为原生且内容无损**。AE 自动化 prefs 注意：脚本跑完让 AE 优雅 `app.quit()` 等几秒再 kill，否则首选项损坏（见 [ae-automation-occlusion-crashstate.md](ae-automation-occlusion-crashstate.md)）。

## Cases
- 2026-06-09 首次：完整 RE + 三层验证；coverage 仅放指针，未将版本门禁表直接融入（用户决策：开本 incident 收口、ldta 线索补进 ldta incident）。
