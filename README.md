# twitter-dore

Twitter のリプ欄でお馴染みの質問テンプレを即座に埋めるための CLI です。YAML テンプレートを読み込み、`{}` で示されたプレースホルダを `promptui` による対話入力で埋め込みます。

## 主な機能

- `twitter-dore run`  
  YAML テンプレートを読み込み、左から順に `{}` を置換します。`{{}}` はリテラルの `{}` として扱われます。
- `twitter-dore new`  
  新規テンプレートを作成します。`--template-inline`/`--template-file` による非対話モードと、`promptui` でフィールドを収集する対話モードを用意しています。
- `twitter-dore version`  
  バージョン情報を表示します。
- `twitter-dore completion <shell>`  
  bash / zsh / fish / PowerShell の補完スクリプトを生成します。

### YAML スキーマ

```yaml
title: <string>
description: <string>
template: |-
  呼び方: {}
  好感度: {}
```

未知のキーは無視されます。`template` が空の場合はエラーとなります。

## 使い方

### テンプレートを実行 (`run`)

```bash
twitter-dore run --in tpl.yaml [--out reply.txt] [--no-empty] [--quiet] [--color=auto|always|never]
```

- `--no-empty` を指定すると、空入力は再入力を求められます。
- `--out` を指定すると UTF-8 でファイル保存します。標準出力は既定で有効、`--quiet` で抑止可能です。
- `--color=auto`（既定）は TTY のときだけ太字 + 下線でプレースホルダ行を強調します。`always` / `never` で明示変更できます。

### テンプレートを作成 (`new`)

```bash
# 非対話モード（必須: --out と --template-inline|--template-file）
twitter-dore new \
  --out tpl.yaml \
  --title "すきなところ" \
  --description "リプで回答するテンプレ" \
  --template-inline "呼び方:{}\n好感度:{}"

# 対話モード（テンプレ本文は行単位で入力、EOF と入力すると終了）
twitter-dore new --out tpl.yaml
```

対話モードでは

1. `title` → `description` → `template line N` の順で `promptui` による入力を行います。
2. テンプレ本文は行ごとに入力し、終了したいタイミングで `EOF` と入力します（空行もそのまま登録できます）。
3. プレースホルダのプレビューは `{}` 部分を強調して `stderr` に表示します。
4. 既存ファイルに上書きする場合は `--force` が必要です。

## ビルド & インストール

```bash
go build -o twitter-dore .
# または Makefile を利用
make build
```

生成されたバイナリを `$PATH` 上の任意の場所に配置してください。

## 開発

- フォーマット: `go fmt ./...`
- テスト: `go test ./...`
- ローカル Lint: `golangci-lint run`
- Makefile:
  - `make fmt` / `make test` / `make lint` / `make build`
  - `make ci` で lint → test → build をまとめて実行

### CI

GitHub Actions（`.github/workflows/ci.yml`）でビルド・テスト・`golangci-lint` を実行します。

### バージョン情報の埋め込み

`go build` 時に次のように指定すると `twitter-dore version` の出力を書き換えられます。

```bash
go build -ldflags "-X github.com/AkatukiSora/twitter-dore/cmd.version=1.2.3 -X github.com/AkatukiSora/twitter-dore/cmd.commit=$(git rev-parse --short HEAD)"
```

## ライセンス

MIT License
