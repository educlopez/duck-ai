# duck-ai

Personal Claude Code toolkit — skills, commands, and setup scripts.
Built for day-to-day work at **Cinetic** (soon Neven).

## Related tools

| Tool | What |
|------|------|
| [ps-lando](http://ps-lando.educalvolopez.com/) | CLI to scaffold a Lando environment with PrestaShop + Panda theme |
| [prestashop-experts](https://github.com/educlopez/prestashop-experts) | Claude agents specialized in PrestaShop development |

## Install

### macOS / Linux — Homebrew

```bash
brew tap educlopez/tap
brew install duck-ai
```

### macOS / Linux — curl

```bash
curl -fsSL https://raw.githubusercontent.com/educlopez/duck-ai/main/install.sh | bash
```

Drops the binary into `~/.local/bin/duck-ai` (override with `DUCK_AI_INSTALL_DIR=/usr/local/bin`). Make sure `~/.local/bin` is on your `PATH`.

Pin a specific version:

```bash
DUCK_AI_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/educlopez/duck-ai/main/install.sh | bash
```

### Windows — Scoop

```powershell
scoop bucket add educlopez https://github.com/educlopez/scoop-bucket
scoop install duck-ai
```

### After install

```bash
duck-ai update    # install Claude/Codex/OpenCode skills + commands
duck-ai doctor    # verify installation
duck-ai           # launch interactive TUI
```

### CLI

```bash
duck-ai                        # interactive TUI
duck-ai install                # same as above
duck-ai install --agent claude # install only to claude (non-interactive)
duck-ai install --all          # install to all detected agents (non-interactive)
duck-ai update                 # re-link skills/commands
duck-ai doctor                 # check symlink health per agent
duck-ai registry               # list installed skills/commands
duck-ai version                # print version
```

The binary auto-detects installed agents (claude, agents, codex, opencode) and
symlinks the appropriate skills/commands directories.

### From source

```bash
git clone git@github.com:educlopez/duck-ai.git
cd duck-ai
go build -o duck-ai .
./duck-ai update
```

## Update

Upgrade the binary, then re-sync skills/commands:

```bash
brew upgrade duck-ai       # macOS / Linux (Homebrew)
# or
scoop update duck-ai       # Windows
# or re-run the curl-pipe installer:
curl -fsSL https://raw.githubusercontent.com/educlopez/duck-ai/main/install.sh | bash

duck-ai update             # re-link skills/commands (backs up conflicts)
```

`duck-ai update --list-backups` and `duck-ai update --restore <ts>` recover prior state if anything went sideways.

## Skills

| Skill | Trigger |
|-------|---------|
| `cinetic-security-setup` | Add Trivy + pnpm supply chain protection to Cinetic GitLab projects |
| `lando-img-placeholder` | Static image placeholder for local Lando/PrestaShop dev |
| `ps-demo-user` | Create demo user in PrestaShop 8 (Lando) |

## Commands

| Command | What |
|---------|------|
| `/lando` | Lando environment helpers |
| `/ps-customer` | Create test customer in PrestaShop |
| `/ps-url` | PrestaShop URL utilities |

## Structure

```
main.go                       Entry point + CLI routing
cmd/
  install.go                  Install (TUI + non-interactive modes)
  update.go                   Update with backup, restore, list-backups
  doctor.go                   Doctor (thin delegate to internal/reports)
  registry.go                 Registry (thin delegate to internal/reports)
internal/
  agents/
    adapter.go                Multi-agent Adapter interface
    claude.go codex.go        Per-agent implementations
    opencode.go generic.go
    registry.go               All() / ByID() factory
  backup/backup.go            Snapshot/Restore + manifest, keep-latest-5 GC
  updater/updater.go          Pure Run(Options) → Report (CLI + TUI share it)
  reports/                    Writer-based doctor + registry renderers
  skillregistry/registry.go   Parse SKILL.md / command frontmatter
  skills/skills.go            Symlink helpers
  tui/
    model.go                  Bubbletea model + screen state machine
    welcome.go                Welcome screen calqued from gentle-ai
    logo.go                   Braille-art duck + yellow gradient
    styles.go                 Lipgloss palette (Rose Pine + duck accents)
skills/                       Claude Code skills (symlinked to ~/.claude/skills/)
claude/commands/              Slash commands (symlinked to ~/.claude/commands/)
.github/workflows/
  release.yml                 Tag push → goreleaser → binaries + brew + scoop
  ci.yml                      go vet + go build on push/PR
  pr-check.yml                400-line PR review-budget check
.goreleaser.yaml              Multi-platform build + brews/scoops publish
install.sh                    Curl-pipe installer (downloads release binary)
doctor.sh                     Bash dependency check (claude CLI, node, pnpm)
docs/                         Notes on adding skills and commands
```

## Adding a skill

See `docs/adding-skills.md`.
