package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/educlopez/duck-ai/internal/backup"
	"github.com/educlopez/duck-ai/internal/updater"
)

// UpdateArgs captures parsed flags for `duck-ai update`.
type UpdateArgs struct {
	DryRun      bool
	AgentID     string
	Yes         bool
	Restore     string
	ListBackups bool
}

func RunUpdate(repoRoot string, args UpdateArgs) error {
	if args.ListBackups {
		return runListBackups()
	}
	if args.Restore != "" {
		return runRestore(args)
	}

	rpt, err := updater.Run(updater.Options{
		RepoRoot: repoRoot,
		DryRun:   args.DryRun,
		AgentID:  args.AgentID,
	})
	if err != nil {
		return err
	}

	if rpt.DryRun {
		fmt.Println("\n  duck-ai update (dry run) — no files were modified.")
	} else {
		fmt.Println("\n  duck-ai update")
	}

	if len(rpt.Agents) == 0 {
		fmt.Println("  No agents detected.")
		return nil
	}

	for _, ar := range rpt.Agents {
		fmt.Printf("\n  Agent: %s\n", ar.Agent)
		for _, it := range ar.Items {
			marker := classMarker(it.Class)
			line := fmt.Sprintf("    %s  %-9s  %s/%s", marker, it.Class, it.Kind, it.Name)
			if it.Err != nil {
				line += "  ERROR: " + it.Err.Error()
			}
			fmt.Println(line)
		}
		fmt.Printf("    summary: noop=%d updated=%d replaced=%d missing=%d failed=%d\n",
			ar.Noop, ar.Updated, ar.Replaced, ar.Missing, ar.Failed)
	}

	if rpt.BackupHits > 0 {
		fmt.Printf("\n  Backup: %d item(s) saved to %s\n", rpt.BackupHits, rpt.BackupDir)
	}
	return nil
}

func runListBackups() error {
	summaries, err := backup.ListBackups()
	if err != nil {
		return err
	}
	fmt.Println("\n  duck-ai backups")
	if len(summaries) == 0 {
		fmt.Println("  (no backups found)")
		return nil
	}
	for _, s := range summaries {
		fmt.Printf("\n  %s\n", s.Timestamp)
		fmt.Printf("    dir:     %s\n", s.Dir)
		fmt.Printf("    entries: %d (%s)\n", s.EntryCount, humanBytes(s.TotalBytes))
		if len(s.ByAgent) > 0 {
			fmt.Printf("    by agent: %s\n", formatByAgent(s.ByAgent))
		}
	}
	return nil
}

func runRestore(args UpdateArgs) error {
	stamp, err := backup.ResolveTimestamp(args.Restore)
	if err != nil {
		return err
	}
	root, err := backup.BackupsRoot()
	if err != nil {
		return err
	}
	batchDir := filepath.Join(root, stamp)
	manifest, err := backup.LoadManifest(batchDir)
	if err != nil {
		return err
	}

	items := backup.PlanRestore(manifest, args.AgentID)
	if len(items) == 0 {
		fmt.Printf("\n  No entries to restore from %s", stamp)
		if args.AgentID != "" {
			fmt.Printf(" for agent %q", args.AgentID)
		}
		fmt.Println()
		return nil
	}

	if args.DryRun {
		fmt.Printf("\n  duck-ai restore (dry run) — backup %s\n", stamp)
	} else {
		fmt.Printf("\n  duck-ai restore — backup %s\n", stamp)
	}

	printRestorePlan(items)

	if args.DryRun {
		return nil
	}

	if !args.Yes && isTTY(os.Stdin) {
		fmt.Print("\n  Apply restore? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.ToLower(strings.TrimSpace(line))
		if line != "y" && line != "yes" {
			fmt.Println("  aborted")
			return nil
		}
	}

	results := backup.ApplyRestore(items)
	fmt.Println("\n  Results:")
	printRestorePlan(results)

	var ok, skipped, failed int
	for _, r := range results {
		switch r.Class {
		case backup.RestoreRestore, backup.RestoreRelink:
			ok++
		case backup.RestoreSkip:
			skipped++
		case backup.RestoreFailed:
			failed++
		}
	}
	fmt.Printf("\n  summary: restored=%d skipped=%d failed=%d\n", ok, skipped, failed)
	return nil
}

func printRestorePlan(items []backup.RestoreItem) {
	for _, it := range items {
		marker := restoreMarker(it.Class)
		line := fmt.Sprintf("    %s  %-7s  %s/%s  %s",
			marker, it.Class, it.Entry.Agent, it.Entry.Kind, it.Entry.OriginalPath)
		if it.Reason != "" {
			line += "  (" + it.Reason + ")"
		}
		if it.Err != nil {
			line += "  ERROR: " + it.Err.Error()
		}
		fmt.Println(line)
	}
}

func restoreMarker(c backup.RestoreClass) string {
	switch c {
	case backup.RestoreRestore:
		return ">"
	case backup.RestoreRelink:
		return "~"
	case backup.RestoreSkip:
		return "!"
	case backup.RestoreFailed:
		return "x"
	}
	return "?"
}

func formatByAgent(by map[string]int) string {
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %d", k, by[k]))
	}
	return strings.Join(parts, ", ")
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func classMarker(c updater.Classification) string {
	switch c {
	case updater.ClassNoop:
		return "~"
	case updater.ClassUpdate:
		return ">"
	case updater.ClassReplace:
		return "!"
	case updater.ClassMissing:
		return "+"
	}
	return "?"
}

// ParseUpdateArgs is a tiny helper kept here so main.go stays light.
func ParseUpdateArgs(args []string) (UpdateArgs, error) {
	var out UpdateArgs
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			out.DryRun = true
		case "--agent":
			if i+1 >= len(args) {
				return out, fmt.Errorf("--agent requires a value")
			}
			out.AgentID = args[i+1]
			i++
		case "--yes":
			out.Yes = true
		case "--restore":
			if i+1 >= len(args) {
				return out, fmt.Errorf("--restore requires a value")
			}
			out.Restore = args[i+1]
			i++
		case "--list-backups":
			out.ListBackups = true
		default:
			if strings.HasPrefix(args[i], "-") {
				_, _ = fmt.Fprintf(os.Stderr, "unknown flag %q\n", args[i])
			}
		}
	}
	if out.Restore != "" && out.ListBackups {
		return out, fmt.Errorf("--restore and --list-backups are mutually exclusive")
	}
	return out, nil
}
