# duck-ai

Personal Claude Code toolkit — skills, commands, and setup scripts.

## Install

```bash
git clone git@github.com:<you>/duck-ai.git
cd duck-ai
chmod +x install.sh doctor.sh
./install.sh
```

## Verify

```bash
./doctor.sh
```

## Update

```bash
git pull
./install.sh   # re-runs; symlinks update automatically
```

## Structure

```
skills/           Claude Code skills (symlinked to ~/.claude-work/skills/)
claude/commands/  Slash commands (symlinked to ~/.claude/commands/)
docs/             Notes on adding skills and commands
```

## Adding a skill

See `docs/adding-skills.md`.
