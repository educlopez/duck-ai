package cmd

import (
	"os"

	"github.com/educlopez/duck-ai/internal/reports"
)

// RegistryArgs is a thin alias of the renderer's args type, kept here so
// main.go's existing call site (`cmd.RegistryArgs`, `cmd.ParseRegistryArgs`)
// continues to compile unchanged.
type RegistryArgs = reports.RegistryArgs

// ParseRegistryArgs delegates to the renderer.
func ParseRegistryArgs(args []string) (RegistryArgs, error) {
	return reports.ParseRegistryArgs(args)
}

// RunRegistry prints the registry report to stdout. The TUI calls
// reports.Registry directly with a captured writer.
func RunRegistry(repoRoot string, args RegistryArgs) error {
	return reports.Registry(os.Stdout, repoRoot, args)
}
