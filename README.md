# gwm

## 対応コマンド

- `gwm create <branch>`
  - 指定ブランチがなければ、`origin/HEAD` が指すデフォルトブランチ（取得できない場合は `main`）から新規作成します。
  - リポジトリ直下の `worktrees/<branch>` に git worktree を追加します。
  - `.gwm/config.json` に登録されたファイルを worktree に展開します。`mode: copy` はファイルコピー、`mode: symlink` はシンボリックリンクで配置します。

- `gwm config add <path> --mode copy|symlink`
  - 管理対象ファイルを設定に追加します。`--mode` 省略時は `copy`。`path` はリポジトリ相対のみ許可され、重複登録はエラーになります。

- `gwm config list`
  - `.gwm/config.json` の内容を JSON で標準出力に表示します。登録が無い場合は `no entries` と表示します。

- `gwm config remove <path>`
  - 登録済みのエントリを削除します。見つからない場合はエラーになります。

- `gwm cd`
  - `git worktree list --porcelain` の結果を元に一覧を表示し、番号入力で移動先を選択します（現在の worktree には `*` マーク）。選択後は対象 worktree に `cd` したシェルを起動します。

## ビルド方法

1. Go 1.25 系を用意します（`go version` で確認）。
2. ルートディレクトリで `go build -o gwm ./cmd/gwm` を実行します。
3. 生成されたバイナリ `./gwm` を任意のパスに配置するか、実行ディレクトリでそのまま利用してください。

## 補足

- 設定は `.gwm/config.json` に JSON で保存されます（存在しない場合は自動作成）。
- 実行例: `go run ./cmd/gwm create feature/foo`、`go run ./cmd/gwm config add path/to/file --mode symlink`。
