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
  - `git worktree list --porcelain` の結果を元に一覧を Bubble Tea UI で表示し、矢印キーまたは数字入力で選択します（現在の worktree には `*` マーク）。
  - 選択後は tmux セッション `gwm-<branch>` に attach（存在しない場合はカレントを `<branch>` で新規作成）。tmux が無い環境では従来どおりシェルを起動します。

- `gwm remove <branch> [--force]`
  - `git worktree remove` で `worktrees/<branch>` を削除します。`--force` を付けると未コミットの変更があっても削除します。
  - 対応する tmux セッションがあれば終了させます（存在しない場合は何もしません）。
  - worktree で実行中のサービスも自動的に停止します。

- `gwm service add <name> --command "..." [--port auto|none|<number>] [--unique]`
  - サービス定義を `.gwm/services.json` に登録します。
  - `--port auto`: 3000-3999 の範囲から未使用ポートを自動割り当て。
  - `--port none`: ポート管理なし。
  - `--port <number>`: 固定ポートを指定。
  - `--unique`: 全 worktree で1インスタンスのみ実行。他の worktree で同名サービスが実行中の場合、停止して現在の worktree で起動し直します。
  - コマンド内で `{port}` プレースホルダーを使うと、実行時に実際のポート番号に置換されます。
    - 例: `gwm service add dev --command "yarn dev --port {port}" --port auto`
    - 例: `gwm service add watcher --command "npm run watch" --port none --unique`

- `gwm service start <name>`
  - 現在の worktree で登録済みサービスを起動します。
  - 専用の tmux セッション `gwm-svc-<worktree>-<name>-p<port>` で実行されます。
  - 固定ポートが他の worktree で使用中の場合、そのサービスを停止して新しく起動します。

- `gwm service stop <name>`
  - 現在の worktree で実行中のサービスを停止します。

- `gwm service list`
  - 全 worktree で実行中のサービス一覧を JSON で表示します。

- `gwm service attach <name>`
  - サービスの tmux セッションに attach してログを確認できます。

- `gwm service definitions`
  - 登録済みのサービス定義一覧を JSON で表示します。

- `gwm service remove <name>`
  - サービス定義を削除します。

## ビルド方法

1. Go 1.25 系を用意します（`go version` で確認）。
2. ルートディレクトリで `go build -o gwm .` を実行します。
3. 生成されたバイナリ `./gwm` を任意のパスに配置するか、実行ディレクトリでそのまま利用してください。

## 補足

- 設定は `.gwm/config.json` に JSON で保存されます（存在しない場合は自動作成）。
- サービス定義は `.gwm/services.json` に JSON で保存されます。
- 実行例: `go run . create feature/foo`、`go run . config add path/to/file --mode symlink`。
