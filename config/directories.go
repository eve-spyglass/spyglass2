package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func (cfg *Config) logDirHint() string {

	switch runtime.GOOS {
	case "windows":
		dir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		fp := filepath.Join(dir, "Documents", "Eve", "logs")
		return fp
	case "linux":
		fallthrough
	case "darwin":
		fallthrough
	default:
		return ""
	}

}

func (cfg *Config) location() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".spyglass", "spyglass_config.json"), err
}
