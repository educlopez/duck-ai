package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Entry describes a single file or directory that was backed up before being
// overwritten by duck-ai.
type Entry struct {
	Agent        string `json:"agent"`
	Kind         string `json:"kind"`
	OriginalPath string `json:"original_path"`
	BackupPath   string `json:"backup_path"`
	Sha256       string `json:"sha256,omitempty"`
}

// Manifest is the JSON manifest written alongside each backup batch.
type Manifest struct {
	Timestamp time.Time `json:"timestamp"`
	Entries   []Entry   `json:"entries"`
}

// Session groups all snapshots taken under a single timestamped backup dir.
type Session struct {
	mu      sync.Mutex
	rootDir string
	stamp   string
	entries []Entry
}

// NewSession opens (lazily) a new backup session under ~/.duck-ai/backups/<RFC3339>.
// The directory is NOT created until the first Snapshot call.
func NewSession() (*Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	stamp := time.Now().UTC().Format("20060102T150405Z")
	root := filepath.Join(home, ".duck-ai", "backups", stamp)
	return &Session{rootDir: root, stamp: stamp}, nil
}

// Root returns the backup directory path. Empty until first snapshot.
func (s *Session) Root() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) == 0 {
		return ""
	}
	return s.rootDir
}

// Count returns how many entries have been snapshotted.
func (s *Session) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

// Snapshot copies the file or directory at originalPath into the session's
// backup root, organized by agent/kind. Symlinks are NOT followed: a symlink
// is recorded with its target string instead of copying through.
func (s *Session) Snapshot(agentID, kind, originalPath string) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, err := os.Lstat(originalPath)
	if err != nil {
		return Entry{}, fmt.Errorf("lstat %s: %w", originalPath, err)
	}

	dstDir := filepath.Join(s.rootDir, agentID, kind)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return Entry{}, fmt.Errorf("mkdir %s: %w", dstDir, err)
	}
	dst := filepath.Join(dstDir, filepath.Base(originalPath))

	entry := Entry{
		Agent:        agentID,
		Kind:         kind,
		OriginalPath: originalPath,
		BackupPath:   dst,
	}

	switch {
	case info.Mode()&os.ModeSymlink != 0:
		target, rerr := os.Readlink(originalPath)
		if rerr != nil {
			return Entry{}, fmt.Errorf("readlink %s: %w", originalPath, rerr)
		}
		if werr := os.WriteFile(dst+".symlink", []byte(target), 0o644); werr != nil {
			return Entry{}, fmt.Errorf("write symlink record: %w", werr)
		}
		entry.BackupPath = dst + ".symlink"

	case info.IsDir():
		if err := copyTree(originalPath, dst); err != nil {
			return Entry{}, err
		}

	default:
		sum, err := copyFile(originalPath, dst)
		if err != nil {
			return Entry{}, err
		}
		entry.Sha256 = sum
	}

	s.entries = append(s.entries, entry)
	return entry, nil
}

// Finalize writes manifest.json and runs the keep-latest-5 garbage collector.
// No-op if no snapshots were taken.
func (s *Session) Finalize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) == 0 {
		return nil
	}

	m := Manifest{Timestamp: time.Now().UTC(), Entries: s.entries}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(s.rootDir, "manifest.json"), data, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	return gcOldBackups(filepath.Dir(s.rootDir), 5)
}

func copyFile(src, dst string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, h), in); err != nil {
		return "", fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyTree(srcRoot, dstRoot string) error {
	return filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(srcRoot, path)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(dstRoot, rel)

		// Lstat to avoid following symlinks inside the tree.
		linfo, lerr := os.Lstat(path)
		if lerr != nil {
			return lerr
		}
		switch {
		case linfo.Mode()&os.ModeSymlink != 0:
			t, rrerr := os.Readlink(path)
			if rrerr != nil {
				return rrerr
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			return os.WriteFile(target+".symlink", []byte(t), 0o644)
		case linfo.IsDir():
			return os.MkdirAll(target, 0o755)
		default:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			_, err := copyFile(path, target)
			return err
		}
	})
}

// BackupsRoot returns the parent directory that holds every backup batch.
func BackupsRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".duck-ai", "backups"), nil
}

// Summary describes a single backup batch on disk.
type Summary struct {
	Timestamp  string
	Dir        string
	EntryCount int
	TotalBytes int64
	ByAgent    map[string]int
}

// ListBackups returns every backup batch under ~/.duck-ai/backups, newest first.
func ListBackups() ([]Summary, error) {
	root, err := BackupsRoot()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var stamps []string
	for _, e := range entries {
		if e.IsDir() {
			stamps = append(stamps, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(stamps)))

	out := make([]Summary, 0, len(stamps))
	for _, stamp := range stamps {
		dir := filepath.Join(root, stamp)
		m, err := LoadManifest(dir)
		s := Summary{Timestamp: stamp, Dir: dir, ByAgent: map[string]int{}}
		if err != nil {
			out = append(out, s)
			continue
		}
		s.EntryCount = len(m.Entries)
		for _, en := range m.Entries {
			s.ByAgent[en.Agent]++
			s.TotalBytes += entryBytes(en.BackupPath)
		}
		out = append(out, s)
	}
	return out, nil
}

// entryBytes returns the on-disk size of a backup entry. For files it is the
// file size; for directories it is the recursive sum of contained file sizes.
// Errors are silently treated as zero — list output should never fail just
// because one stale entry on disk is unreadable.
func entryBytes(path string) int64 {
	info, err := os.Lstat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}
	var total int64
	_ = filepath.Walk(path, func(_ string, fi os.FileInfo, werr error) error {
		if werr != nil || fi == nil {
			return nil
		}
		if !fi.IsDir() {
			total += fi.Size()
		}
		return nil
	})
	return total
}

// ResolveTimestamp accepts either a full stamp or a unique prefix and returns
// the canonical timestamp directory name.
func ResolveTimestamp(prefix string) (string, error) {
	root, err := BackupsRoot()
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no backups found")
		}
		return "", err
	}

	var matches []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == prefix {
			return e.Name(), nil
		}
		if strings.HasPrefix(e.Name(), prefix) {
			matches = append(matches, e.Name())
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no backup matches %q", prefix)
	case 1:
		return matches[0], nil
	default:
		sort.Strings(matches)
		return "", fmt.Errorf("ambiguous prefix: matches %s", strings.Join(matches, ", "))
	}
}

// LoadManifest reads manifest.json from a backup batch directory.
func LoadManifest(dir string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

// RestoreClass classifies what a restore would do for a single entry.
type RestoreClass string

const (
	RestoreRestore RestoreClass = "restore" // copy backup over the target
	RestoreRelink  RestoreClass = "relink"  // target is a duck-ai symlink; safe to drop
	RestoreSkip    RestoreClass = "skip"    // target was modified by user; refuse to clobber
	RestoreFailed  RestoreClass = "failed"  // restore attempted but failed (sha mismatch, IO)
)

// RestoreItem is the per-entry result of restore planning or execution.
type RestoreItem struct {
	Entry  Entry
	Class  RestoreClass
	Reason string
	Err    error
}

// PlanRestore walks the manifest and classifies each entry against the current
// state of the filesystem. It performs no mutation.
func PlanRestore(m *Manifest, agentFilter string) []RestoreItem {
	out := make([]RestoreItem, 0, len(m.Entries))
	for _, e := range m.Entries {
		if agentFilter != "" && e.Agent != agentFilter {
			continue
		}
		item := RestoreItem{Entry: e, Class: RestoreRestore}

		info, lerr := os.Lstat(e.OriginalPath)
		switch {
		case lerr != nil && os.IsNotExist(lerr):
			// target gone — just restore
		case lerr != nil:
			item.Class = RestoreFailed
			item.Err = lerr
			item.Reason = "lstat failed"
		case info.Mode()&os.ModeSymlink != 0:
			item.Class = RestoreRelink
			item.Reason = "target is symlink"
		default:
			item.Class = RestoreSkip
			item.Reason = "target modified"
		}

		out = append(out, item)
	}
	return out
}

// ApplyRestore executes a previously-planned restore. Items already marked
// RestoreSkip or RestoreFailed are passed through unchanged.
func ApplyRestore(items []RestoreItem) []RestoreItem {
	for i := range items {
		it := &items[i]
		if it.Class == RestoreSkip || it.Class == RestoreFailed {
			continue
		}
		if it.Class == RestoreRelink {
			if err := os.Remove(it.Entry.OriginalPath); err != nil && !os.IsNotExist(err) {
				it.Class = RestoreFailed
				it.Err = fmt.Errorf("remove symlink: %w", err)
				continue
			}
		}
		if err := restoreOne(it.Entry); err != nil {
			it.Class = RestoreFailed
			it.Err = err
			continue
		}
		it.Class = RestoreRestore
	}
	return items
}

func restoreOne(e Entry) error {
	if strings.HasSuffix(e.BackupPath, ".symlink") {
		target, err := os.ReadFile(e.BackupPath)
		if err != nil {
			return fmt.Errorf("read symlink record: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(e.OriginalPath), 0o755); err != nil {
			return err
		}
		return os.Symlink(strings.TrimSpace(string(target)), e.OriginalPath)
	}

	info, err := os.Lstat(e.BackupPath)
	if err != nil {
		return fmt.Errorf("lstat backup: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(e.OriginalPath), 0o755); err != nil {
		return err
	}
	if info.IsDir() {
		return copyTree(e.BackupPath, e.OriginalPath)
	}

	sum, err := copyFileWithMode(e.BackupPath, e.OriginalPath, info.Mode())
	if err != nil {
		return err
	}
	if e.Sha256 != "" && sum != e.Sha256 {
		return fmt.Errorf("sha mismatch: got %s want %s", sum, e.Sha256)
	}
	return nil
}

func copyFileWithMode(src, dst string, mode os.FileMode) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
	if err != nil {
		return "", fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, h), in); err != nil {
		return "", fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func gcOldBackups(parent string, keep int) error {
	entries, err := os.ReadDir(parent)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) <= keep {
		return nil
	}
	sort.Strings(dirs)
	for _, d := range dirs[:len(dirs)-keep] {
		_ = os.RemoveAll(filepath.Join(parent, d))
	}
	return nil
}
