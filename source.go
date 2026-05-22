package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// materializeEmbeddedSource extracts the embedded skills/ and claude/ trees
// into a per-version cache directory and returns its absolute path.
//
// Layout:
//
//	~/.duck-ai/source/<version>/skills/...
//	~/.duck-ai/source/<version>/claude/commands/...
//	~/.duck-ai/source/<version>/.version    (marker)
//
// Re-extraction is skipped when a `.version` marker exists with matching
// content. If $HOME is unavailable, falls back to $TMPDIR/duck-ai/source/.
func materializeEmbeddedSource(version string) (string, error) {
	base, err := sourceCacheBase()
	if err != nil {
		return "", err
	}
	target := filepath.Join(base, version)

	// Fast path: already materialized for this version.
	if marker, err := os.ReadFile(filepath.Join(target, ".version")); err == nil {
		if string(marker) == version {
			return target, nil
		}
	}

	// Wipe and recreate so partial/corrupt extractions don't linger.
	if err := os.RemoveAll(target); err != nil {
		return "", fmt.Errorf("clean cache dir: %w", err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	err = fs.WalkDir(embeddedSource, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		dest := filepath.Join(target, path)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := embeddedSource.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Marker last so we never claim a partial extraction is complete.
	if err := os.WriteFile(filepath.Join(target, ".version"), []byte(version), 0o644); err != nil {
		return "", fmt.Errorf("write version marker: %w", err)
	}
	return target, nil
}

func sourceCacheBase() (string, error) {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".duck-ai", "source"), nil
	}
	return filepath.Join(os.TempDir(), "duck-ai", "source"), nil
}
