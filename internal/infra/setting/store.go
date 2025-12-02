package setting

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/example/gwm/internal/domain"
)

// Load はリポジトリ直下の .gwm/setting.json を読み込む。
// ファイルが無い場合や空の場合はデフォルト設定を返す。
func Load(repoDir string) (domain.Settings, error) {
	path := filepath.Join(repoDir, ".gwm", "setting.json")

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.DefaultSettings(), nil
	}
	if err != nil {
		return domain.DefaultSettings(), err
	}

	if len(data) == 0 {
		return domain.DefaultSettings(), nil
	}

	var settings domain.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return domain.DefaultSettings(), err
	}

	return settings, nil
}
