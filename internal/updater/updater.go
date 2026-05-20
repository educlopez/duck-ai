package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/educlopez/duck-ai/internal/agents"
	"github.com/educlopez/duck-ai/internal/backup"
	"github.com/educlopez/duck-ai/internal/skills"
)

// Classification of a single target path during update planning.
type Classification string

const (
	ClassNoop    Classification = "noop"    // symlink already points at the right source
	ClassUpdate  Classification = "update"  // symlink pointing somewhere else / stale
	ClassReplace Classification = "replace" // regular file or dir; must be backed up
	ClassMissing Classification = "missing" // does not exist; create symlink
)

// PlanItem describes one source item and what update would do at its target.
type PlanItem struct {
	Agent  string
	Kind   string // "skills" | "commands"
	Name   string
	Src    string
	Dst    string
	Class  Classification
	Note   string
	Err    error
}

// AgentReport aggregates plan and execution results for one agent.
type AgentReport struct {
	Agent    string
	Items    []PlanItem
	Noop     int
	Updated  int
	Replaced int
	Missing  int
	Failed   int
}

// Report is the full update result.
type Report struct {
	DryRun     bool
	BackupDir  string
	BackupHits int
	Agents     []AgentReport
}

// Options controls the update run.
type Options struct {
	RepoRoot string
	DryRun   bool
	AgentID  string // empty = all detected
}

// Run executes the update flow. Pure: the caller decides how to render the Report.
func Run(opts Options) (*Report, error) {
	absRepo, err := filepath.Abs(opts.RepoRoot)
	if err != nil {
		absRepo = opts.RepoRoot
	}

	targets, err := selectAgents(opts.AgentID)
	if err != nil {
		return nil, err
	}

	allSkills, err := skills.DiscoverSkills(absRepo)
	if err != nil {
		return nil, err
	}
	allCommands, err := skills.DiscoverCommands(absRepo)
	if err != nil {
		return nil, err
	}

	rpt := &Report{DryRun: opts.DryRun}

	var session *backup.Session
	if !opts.DryRun {
		session, err = backup.NewSession()
		if err != nil {
			return nil, err
		}
	}

	for _, a := range targets {
		if !a.Detect() {
			continue
		}
		ar := AgentReport{Agent: a.ID()}

		ar.Items = append(ar.Items, planFor(a, allSkills, "skills", a.SkillsDir())...)
		ar.Items = append(ar.Items, planFor(a, allCommands, "commands", a.CommandsDir())...)

		if !opts.DryRun {
			for i := range ar.Items {
				applyItem(&ar.Items[i], session)
			}
		}

		for _, it := range ar.Items {
			switch it.Class {
			case ClassNoop:
				ar.Noop++
			case ClassUpdate:
				ar.Updated++
			case ClassReplace:
				ar.Replaced++
			case ClassMissing:
				ar.Missing++
			}
			if it.Err != nil {
				ar.Failed++
			}
		}
		rpt.Agents = append(rpt.Agents, ar)
	}

	if session != nil {
		if err := session.Finalize(); err != nil {
			return rpt, err
		}
		rpt.BackupDir = session.Root()
		rpt.BackupHits = session.Count()
	}

	return rpt, nil
}

func selectAgents(id string) ([]agents.Adapter, error) {
	if id == "" {
		return agents.Detected(), nil
	}
	a, ok := agents.ByID(id)
	if !ok {
		return nil, fmt.Errorf("unknown agent %q", id)
	}
	return []agents.Adapter{a}, nil
}

func planFor(a agents.Adapter, items []skills.Skill, kind, dstDir string) []PlanItem {
	if dstDir == "" || len(items) == 0 {
		return nil
	}
	out := make([]PlanItem, 0, len(items))
	for _, s := range items {
		dst := filepath.Join(dstDir, s.Name)
		out = append(out, PlanItem{
			Agent: a.ID(),
			Kind:  kind,
			Name:  s.Name,
			Src:   s.SrcPath,
			Dst:   dst,
			Class: classify(s.SrcPath, dst),
		})
	}
	return out
}

func classify(src, dst string) Classification {
	info, err := os.Lstat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return ClassMissing
		}
		return ClassReplace
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(dst)
		if err != nil {
			return ClassUpdate
		}
		if target == src {
			return ClassNoop
		}
		return ClassUpdate
	}
	return ClassReplace
}

func applyItem(it *PlanItem, session *backup.Session) {
	switch it.Class {
	case ClassNoop:
		return
	case ClassMissing:
		if err := ensureLink(it.Src, it.Dst); err != nil {
			it.Err = err
		}
	case ClassUpdate:
		if err := os.Remove(it.Dst); err != nil {
			it.Err = fmt.Errorf("remove stale symlink: %w", err)
			return
		}
		if err := ensureLink(it.Src, it.Dst); err != nil {
			it.Err = err
		}
	case ClassReplace:
		if _, err := session.Snapshot(it.Agent, it.Kind, it.Dst); err != nil {
			it.Err = fmt.Errorf("backup: %w", err)
			return
		}
		if err := os.RemoveAll(it.Dst); err != nil {
			it.Err = fmt.Errorf("remove original: %w", err)
			return
		}
		if err := ensureLink(it.Src, it.Dst); err != nil {
			it.Err = err
		}
	}
}

func ensureLink(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
	}
	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", dst, src, err)
	}
	return nil
}
