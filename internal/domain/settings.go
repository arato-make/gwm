package domain

// Settings は gwm の起動時に読み込むユーザー設定。
// フィールドを増やした際もゼロ値で安全に扱えるようにする。
type Settings struct {
}

// DefaultSettings は設定ファイルが存在しない場合に利用するデフォルト値。
func DefaultSettings() Settings {
	return Settings{}
}
