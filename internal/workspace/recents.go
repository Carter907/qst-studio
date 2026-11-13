package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func configPath(filename string) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	dir := filepath.Join(configDir, "qst-studio")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, filename)
}

// GetRecentGraphs returns previously opened graph directory paths that still exist on disk.
func GetRecentGraphs() ([]string, error) {
	data, err := os.ReadFile(configPath("recents.json"))
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}

	var recents []string
	if err := json.Unmarshal(data, &recents); err != nil {
		return []string{}, nil
	}

	existing := make([]string, 0, len(recents))
	for _, p := range recents {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			existing = append(existing, p)
		}
	}
	return existing, nil
}

// SaveRecentGraph persists a directory at the top of the recent graphs list (max 10).
func SaveRecentGraph(dirPath string) error {
	if dirPath == "" {
		return nil
	}
	recents, _ := GetRecentGraphs()
	updated := []string{dirPath}
	for _, p := range recents {
		if p != dirPath && len(updated) < 10 {
			updated = append(updated, p)
		}
	}
	data, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath("recents.json"), data, 0644)
}
