# ⚠ 外部 AE 版本降级器 CEP 工具 RE

外部 AE 版本降级器 CEP 工具 RE

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
6. **主流 2025→2020 路径会静默损坏「非相邻显式 track-matte」绑定**（2026-06-22 实证，详下节）。砍掉的 ldta 尾 4 字节 = `@0xA0` track-matte 显式源 ID（AE23+），**不是恒零 padding**。用了 track matte（老功能、2020 支持）且源层非相邻的真实工程，降级后遮罩源被抹、AE 2020 按隐式「相邻上层」规则重绑到错误层——无报错。**这是 bug #1–#5 之外唯一在「主流路径 + 纯老特性内容」下就触发的内容损坏**（#1/#2/#5 是边缘 target/输入护栏误杀，#4 是 2018/2019 偏门档）。

## 验证方法（三层互证，无矛盾）

- 静态：node `vm` stub `require`+补 `Buffer`，只跑纯查表解码器反混淆（不调 convertFile），逐函数读出偏移/阈值/护栏。
- live：同 vm 真 fs/path 调真 `convertFile` 对 fixture 副本转换，自写字节 walker diff —— 朴素偏移 `head.dataStart+[1,3,4,5,6,7]`+write-if-different、trimmedLdta 164→160、comr/CIF3/OvdG 删增、bug#1/#2 全部活证。
- ship-gate：`selection_both_layers`(AE2025, 2 图层)→2020 经 `scripts/ae-worker/ae_run.ps1` 开 AE 2020 → 4 items / Comp 1 / 2 层同名同序，无"损坏/跳过"，**工具主流路径(2025→2020)产出 AE 视为原生且内容无损**。AE 自动化 prefs 注意：脚本跑完让 AE 优雅 `app.quit()` 等几秒再 kill，否则首选项损坏（见 [ae-automation-occlusion-crashstate.md](ae-automation-occlusion-crashstate.md)）。

## 实证：非相邻显式 track-matte 降级后静默丢绑（bug #6，2026-06-22）

把 bug #6 从「逻辑必然推断」升级成**手上可复现的实证**。降级器源码不在磁盘（外部 CEP），故忠实**复现其主流 2025→2020 动作**喂真 AE 2020。

**攻击输入（纯老特性，本仓 API 从零造）**：comp `MATTE` 4 个全屏 solid，层序 top→bottom = `[M, X, Y, T]`；`T.SetTrackMatteSource(M, TrackMatteLuma)` —— 源 M 在最顶、被遮层 T 在最底，**非相邻（中隔 X、Y）**。这正是 `@0xA0` 显式源 ID 存在的理由（2020 隐式规则只能表达「相邻上层」）。luma 取值刻意区分：M=白、Y=黑，使渲染也可分辨。

**复现的降级三步**（`rifx.Parse` → 改 → `Chunk.Write` 自动重算 LIST size）：
1. `trimmedLdta` 164→160：每个 164-ldta 砍尾 4 字节（丢 `@0xA0..0xA3`）。
2. head 指纹 2020：把 `head.Data` 的 byte[1],[3],[4],[5],[6],[7] 从真实原生 2020 文件（`test_data/fixtures/renderer_ae2020_r0.aep`）拷过来（`5d/_/16/0b/0b/86/2d`，对上节版本指纹表 2020 行）。
3. 删 AE2022+ 门禁 chunk `pdvc`（本例无 `ppSn`）。

**Go 侧字节实证**（同一 parser 读前后）：
- BEFORE（AE2025，164-ldta）：`T.TrackMatteLayerID = 14 = M.ID` —— **本仓 `SetTrackMatteSource` 正确写 @0xA0，无 0xA0=0 写 bug**（顺手排除了「我们自己也假设 0xA0=0」的疑虑）。`T.TrackMatte(@0x6B)=3`（LUMA）。
- AFTER（trim 后）：`T.TrackMatteLayerID = 0`（显式源**抹除**），`T.TrackMatte(@0x6B)=3` **存活**（孤儿：模式在、源没了）。

**AE 2020 实测**（v17.7，`scripts/ae-worker/ae_run.ps1` 无人值守开全降级文件 dump DOM）：
```
OPENED silently (no corruption dialog)        ← 静默，无报错
idx=1 M trackMatteType=NONE enabled=true       ← 原显式源 M 沦为普通可见层
idx=2 X trackMatteType=NONE enabled=true
idx=3 Y trackMatteType=NONE enabled=true
idx=4 T trackMatteType=LUMA enabled=true        ← T 仍 luma 遮罩，AE2020 隐式源 = 相邻上层 Y（非 M）
```
结论坐实：遮罩**模式存活、显式源丢失** → AE 2020 把 T 的遮罩从「非相邻的 M」静默重绑到「相邻的 Y」，**无任何报错**。设计师天天用的 track matte + 任意层当源（非相邻）= 此工具主流路径的内容损坏硬伤。

**边界诚实**：复现的是降级器**已 RE 的主流路径动作**（非跑真工具，源码不在盘）；故实测证明的是「`@0xA0` 被 trim + 戳成 2020 的文件在 AE 2020 里遮罩重绑」，叠加上节「真工具确做 trimmedLdta 164→160」的 RE 事实，构成完整链条。`enabled=true` on Y 是次要观察（AE 开档未自动关被遮源的 video），不影响主结论。探针 `tmp_debug/downgrade_probe/`（gen + verify.jsx）一次性，findings 落此后删。

## Cases
- 2026-06-09 首次：完整 RE + 三层验证；coverage 仅放指针，未将版本门禁表直接融入（用户决策：开本 incident 收口、ldta 线索补进 ldta incident）。
- 2026-06-22 bug #6 实证：非相邻显式 track-matte 工程经复现的主流 2025→2020 降级后，AE 2020 静默把遮罩源 M→相邻 Y 重绑（DOM dump 实测）；顺带实证本仓 `SetTrackMatteSource` 写 @0xA0 正确。
