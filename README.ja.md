# aep-parser

**After Effects を起動せずに、Go で `.aep` プロジェクトを作成・編集・出力。**

[中文](README.md) · [English](README.en.md) · **日本語**

## 数行のコードから、アニメーション付きの AE プロジェクトへ

以下は[実行可能なサンプル](examples/authoring/main.go)の主要部分です。このリポジトリ内で `internal/aep` の作成・編集 API を使って実行します。`must` / `check` はサンプルで定義するエラー処理用の関数です。定義も下に掲載しています。

### プロジェクトを作成し、テキストと平面レイヤーを追加

```go
import "github.com/yueli-fx/aep-parser/internal/aep"

project := aep.NewProject(aep.TargetAE2020)
comp := must(aep.NewComposition(project, "Hello AEP", 1920, 1080, 30, 5))

must(aep.NewTextLayer(comp, "Title"))
must(aep.NewSolidLayer(comp, "Card", 640, 360, [3]float64{0.1, 0.8, 0.7}))
must(aep.NewSolidLayer(comp, "Draft", 100, 100, [3]float64{1, 0, 0}))
```

これで **1080p・30 fps・5 秒**のコンポジションと、3 つのレイヤーができました。続けてテキストを編集し、エフェクトを追加します。

`must` は `(戻り値, error)` を受け取り、エラーを確認してから戻り値を返します。`check` は `error` のみを返す操作に使います。どちらも Go の組み込み関数やライブラリの API ではありません。`main` の外に定義します。

```go
// エラーが起きたら処理を止め、不正な状態で書き込みを続けない。
func must[T any](value T, err error) T {
    check(err)
    return value
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}
```

### ガウスブラーを追加し、テキストとパラメーターを変更

```go
// メモリ上で再解析し、新規レイヤーを編集可能な状態にする。AE の起動は不要。
project = must(aep.Reopen(project))
comp = project.Compositions[0]
check(comp.LayerByName("Title").SetText("HELLO, AEP"))

card := comp.LayerByName("Card")
blur := must(aep.AddEffect(card, aep.EffectGaussianBlur))
must(aep.SetEffectParam(card, blur, "Blurriness", 30.0))
```

ブラーの量を **30** に設定しました。パラメーターには `"Blurriness"` や `"模糊度"` のような英語・中国語の名前を指定すると、ライブラリが内部識別子を解決します。この作成・編集 API はパラメーター名のみを受け付け、内部の match-name は受け付けません。日本語のパラメーター名の対応は保証していないため、この例では英語名を使います。

ガウスブラーは[追加可能な 231 種類のエフェクトテンプレート](internal/serializer/mutate_effect_add.go)のひとつです。[エフェクトのサンプル](showcase/effects/gen.go)では、シャドウ、ストローク、色調補正などの設定も確認できます。

### ブラーをアニメーションさせ、下書きレイヤーを削除して保存

```go
// 最初の 1 秒でブラーの量を 30 から 0 にする。
must(aep.AnimateEffectParam(card, blur, "Blurriness",
    []aep.ScalarKeyframe{{Time: 0, Value: 30}, {Time: 1, Value: 0}}))

for index, layer := range comp.Layers {
    if layer.Name == "Draft" {
        check(aep.DeleteLayer(comp, index)) // インデックスは 0 始まり。ここでは平面を削除。
        break
    }
}

file := must(os.Create("hello.aep"))
check(project.WriteAEP(file))
check(file.Close())
```

テキスト、平面、ブラーのキーフレームを含む `.aep` ができあがります。**作成・編集・ファイル出力には AE のインストールも不要です。** 画面のレンダリングやその後の編集には AE を使います。完全版のプログラムには出力の再解析チェックがあり、既存ファイルの上書きも防ぎます。

### 実行してみる

```sh
git clone https://github.com/yueli-fx/aep-parser.git
cd aep-parser
go run ./examples/authoring -out tmp/hello.aep
go run ./cmd/aep inspect -in tmp/hello.aep
```

Go 1.25.13、または同系列の新しいパッチバージョンが必要です。上記の作成・編集 API はリポジトリ内のプログラム向けです。自分の Go モジュールに組み込む場合は、[公開 SDK のサンプル](examples/sdk/README.md)にある `Open`、`Export`、`Compile` などを使ってください。

## ほかにも、こんなプロジェクトを生成できます

- **[プロシージャルな炎](showcase/procedural-fx/gen.go)**：3 層のノイズ、温度を表す配色、ぼかしたマスク、加算合成、Glow を組み合わせた炎のアニメーション。[検証記録](showcase/procedural-fx/INDEX.md)
- **[3D カメラシーン](showcase/3d-camera/gen.go)**：カメラと奥行きの異なるカードを作成し、Z 方向の視差と Y 軸回転による遠近感を表現。[検証記録](showcase/3d-camera/INDEX.md)
- **[シェイプアニメーション](showcase/shape-filters/INDEX.md)**：Trim Paths、Repeater、ZigZag、Wiggle など 11 系統のフィルターとその組み合わせ。
- **[エクスプレッションによるアニメーション](showcase/expressions/INDEX.md)**：レイヤー間の参照、loopOut、wiggle、スライダーによる制御。

[全 25 件の showcase →](showcase/README.md) · [ローワーサード、タイトルアニメーション、プログレスバーのレシピ →](examples/projects/README.md)

JSON レシピから直接生成することもできます。

```sh
go run ./cmd/aeprecipe compile -recipe examples/projects/animated-title.json -out tmp/title.aep -json
```

## 既存プロジェクトの解析・比較・バージョン移行

```sh
# コンポジション、レイヤー、エフェクト、キーフレームなどを取得
go run ./cmd/aep profile -in project.aep

# 編集前後のプロジェクトを比較
go run ./cmd/aep diff -expected before.aep -actual after.aep

# 対応するプロジェクト構造を AE2025 向けに変換
go run ./cmd/aep migrate -in source.aep -target AE2025 -out tmp/migrated.aep
```

一括調査、プロパティのスナップショット取得、変更対象外のデータを保持した数値プロパティの編集は、[5 つの公開 SDK サンプル](examples/sdk/README.md)を参照してください。複雑なプロジェクトの作り方を調べるには、レイヤーごとのエフェクト構成、再現手順、学習課題をまとめた [HTML 技法レポート](docs/self-hosted-reports.md)を生成できます。

## 実装の規模と検証状況

**関数・メソッドの機能項目 500 件 · エフェクトテンプレート 231 種類 · 16 分野 · AE2020–AE2025 の 6 バージョンへの書き込み。**

RIFX コンテナー、入れ子のチャンク、バイト配置から、レイヤー構造、2D/3D 変換、テキストスタイル、シェイプとグラデーション、キーフレームとイージング、エクスプレッション、エフェクト、マスクまで実装しています。解析で得た知識や失敗例は **151 本の公開ノート**にまとめ、個別機能を試せる **151 件の Recipe サンプル**も用意しています。

集計元は[機能インデックス](docs/capabilities.json)と[エフェクト登録表](internal/serializer/mutate_effect_add.go)です。382 メソッド + 118 関数のうち、438 項目が stable、62 項目が alpha。検証区分は ae-accept が 266 項目、render-pixel が 163 項目です。これらは内部実装の機能と既存の検証記録の集計です。個別のインターフェースと制限は[機能一覧](docs/capabilities.md)を参照してください。[バージョン移行の回帰検証記録](registry/evidence/versioned-aep-migration/migration_matrix_verify/smoke_all/ledger.md)には 906 通りの組み合わせがあり、成功 900、スキップ 6、失敗 0 です。

## ドキュメント

| 目的 | 入口 |
| --- | --- |
| Go / HTTP から利用する | [利用ガイド（英語）](docs/usage.en.md) · [SDK サンプル](examples/sdk/README.md) |
| JSON からプロジェクトを生成する | [Recipe サンプル](examples/recipes) · [フィールドリファレンス](docs/recipe.md) · [Schema](docs/recipe_schema.json) |
| AEP の内部構造を調べる | [ナレッジベース](docs/knowledge/README.md) · [API リファレンス](docs/README.md) |
| 対応範囲と互換性を確認する | [機能一覧](docs/capabilities.md) · [利用方法と制限（英語）](docs/usage.en.md) |
| 開発に参加する | [貢献ガイド](CONTRIBUTING.md) · [検証手順](docs/knowledge/workflow/verify.md) · [公開前の確認記録](docs/open-source-audit.md) |

現在はプレリリース段階です。解析と生成のコアは Pure Go で実装しており、Windows、macOS、Linux でビルドできます。ナレッジノートは主に中国語、生成された API / Recipe リファレンスは主に英語です。サンプルごとの AE 上での検証状況は各記録を参照してください。[AE2022 のパラメーター名サンプル](examples/authoring/ae2022-named-effects/README.md)は、生成・再解析の検証に加えて、ユーザーから実機テスト成功の報告を受けています。

## ライセンス

**[PolyForm Noncommercial 1.0.0](LICENSE)**。ライセンスで許可された非商用利用は無料です。その許可範囲外での商用利用、販売、商用製品への組み込みには、別途書面による商用ライセンスが必要です。

本プロジェクトは **source-available（ソースコード公開型）** です。無料で利用できる範囲、組織に関する例外、商用ライセンスの相談方法、生成コンテンツの扱いは[利用と商用ライセンスの説明（英語）](LICENSING.md#english)を参照してください。第三者の権利表示は [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)、セキュリティに関する報告方法は [SECURITY.md](SECURITY.md) に記載しています。

本プロジェクトは Adobe と提携していません。Adobe、After Effects、および関連する商標はそれぞれの権利者に帰属します。
