# duck-ai

Personal Claude Code toolkit — skills, commands, and setup scripts.
Built for day-to-day work at **Cinetic** (soon Neven).

## Related tools

| Tool | What |
|------|------|
| [ps-lando](http://ps-lando.educalvolopez.com/) | CLI to scaffold a Lando environment with PrestaShop + Panda theme |
| [prestashop-experts](https://github.com/educlopez/prestashop-experts) | Claude agents specialized in PrestaShop development |

## Install

```bash
git clone git@github.com:educlopez/duck-ai.git
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
