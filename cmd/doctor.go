package cmd

import (
	"os"

	"github.com/educlopez/duck-ai/internal/reports"
)

// RunDoctor prints the doctor report to stdout. The TUI calls
// reports.Doctor directly with a captured writer.
func RunDoctor(repoRoot string) error {
	return reports.Doctor(os.Stdout, repoRoot)
}
