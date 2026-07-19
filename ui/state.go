package ui

import (
	"os"
	"path/filepath"
	"strings"
)

const stateFileRelPath = "deckline/last_directory"

// defaultBrowseDirectory is where the file picker starts when there's no
// saved state yet: the user's home, not the process's working directory
// (which is often wherever `go run`/the binary happened to be launched from).
func defaultBrowseDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

// loadLastDirectory returns the last directory browsed in a previous run,
// falling back to defaultBrowseDirectory if there's no saved state or it no
// longer exists. Best-effort: any failure to read state just means
// starting fresh.
func loadLastDirectory() string {
	path, err := statePath()
	if err != nil {
		return defaultBrowseDirectory()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultBrowseDirectory()
	}
	dir := strings.TrimSpace(string(data))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return defaultBrowseDirectory()
	}
	return dir
}

// saveLastDirectory persists dir for the next run. Best-effort: failures
// are silently ignored, this is a convenience, not a critical path.
func saveLastDirectory(dir string) {
	path, err := statePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(dir), 0o644)
}

// statePath lives under the user's cache directory, not config: it's a
// disposable convenience (which folder you were last browsing), not
// something worth backing up or hand-editing like real configuration.
func statePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, stateFileRelPath), nil
}
