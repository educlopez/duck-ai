package main

import "embed"

// embeddedSource bundles the duck-ai source tree (skills + claude commands)
// directly into the binary so curl-pipe installs (which land the binary in
// ~/.local/bin/ with no sibling repo) still have something to symlink from.
//
// The `all:` prefix ensures that files starting with "_" or "." are also
// embedded — paranoia in case a SKILL.md ever picks up a sibling like
// "_meta.md" or ".version".
//
//go:embed all:skills all:claude
var embeddedSource embed.FS
