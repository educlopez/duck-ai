# duck-ai

Personal Claude Code toolkit — skills, commands, and setup scripts.
Built for day-to-day work at **Cinetic** (soon Neven).

## Related tools

| Tool | What |
|------|------|
| [ps-lando](http://ps-lando.educalvolopez.com/) | CLI to scaffold a Lando environment with PrestaShop + Panda theme |
| [prestashop-experts](https://github.com/educlopez/prestashop-experts) | Claude agents specialized in PrestaShop development |

## Install

### Option A — Go binary (recommended)

Requires Go 1.22+.

```bash
git clone git@github.com:educlopez/duck-ai.git
cd duck-ai
go build -o duck-ai .
./duck-ai          # launches interactive TUI
```

Or install to `$GOPATH/bin` for global access:

```bash
go install github.com/educlopez/duck-ai@latest
duck-ai
```

#### CLI flags

```bash
duck-ai                        # interactive TUI
duck-ai install                # same as above
duck-ai install --agent claude # install only to claude (non-interactive)
duck-ai install --all          # install to all detected agents (non-interactive)
duck-ai doctor                 # check symlink health per agent
duck-ai version                # print version
```

The binary auto-detects installed agents (claude, agents, codex, opencode) and
symlinks the appropriate skills/commands directories. Set `DUCK_AI_DIR` to
override the repo root location.

### Option B — Bash fallback

```bash
git clone git@github.com:educlopez/duck-ai.git
cd duck-ai
chmod +x install.sh doctor.sh
./install.sh
```

## Verify

```bash
duck-ai doctor   # Go binary
# or
./doctor.sh      # bash fallback
```

## Update

```bash
git pull
duck-ai install   # re-runs; symlinks update automatically
```

## Skills

| Skill | Trigger |
|-------|---------|
| `cinetic-security-setup` | Add Trivy + pnpm supply chain protection to Cinetic GitLab projects |
| `gitlab-security-setup` | Generic GitLab dependency scanning (non-Cinetic projects) |
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
main.go              Entry point + CLI routing
cmd/
  install.go         Install command (TUI + non-interactive modes)
  doctor.go          Doctor command
internal/
  agents/agents.go   Agent definitions + detection logic
  skills/skills.go   Skill discovery + symlink operations
  tui/model.go       Bubbletea TUI model (5-screen flow)
  tui/styles.go      Lipgloss styles + banner
skills/              Claude Code skills (symlinked to ~/.claude/skills/)
claude/commands/     Slash commands (symlinked to ~/.claude/commands/)
install.sh           Bash fallback installer
doctor.sh            Bash fallback doctor
docs/                Notes on adding skills and commands
```

## Adding a skill

See `docs/adding-skills.md`.
