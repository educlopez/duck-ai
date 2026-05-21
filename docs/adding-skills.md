# Adding a skill

1. Create `skills/<skill-name>/SKILL.md` in this repo
2. Frontmatter required:
   ```yaml
   ---
   name: skill-name
   description: >
     When to trigger this skill. Be specific — Claude uses this to decide
     whether to invoke the skill. Include contexts, project types, keywords.
   version: "0.1.0"
   ---
   ```
3. Locally (during development):
   ```bash
   go build -o duck-ai .
   ./duck-ai update           # re-link skills into ~/.claude/skills (with backup)
   ./duck-ai doctor           # confirm the new skill appears under "managed"
   ./duck-ai registry         # check version + drift status
   ```
4. Ship to the team:
   ```bash
   git add skills/<skill-name>/
   git commit -m "feat: add <skill-name> skill"
   git push
   git tag vX.Y.Z
   git push origin vX.Y.Z      # GH Actions builds + publishes Homebrew/Scoop/binaries
   ```
   Team members pick it up with `brew upgrade duck-ai` (or `scoop update duck-ai`) followed by `duck-ai update`.

## Adding a command

1. Create `claude/commands/<command-name>.md`
2. Frontmatter required:
   ```yaml
   ---
   name: command-name
   description: One-line summary of what the command does
   version: "0.1.0"
   ---
   ```
3. Available in Claude Code as `/<command-name>` after running `duck-ai update`.

## Skill scope

| Scope | Where |
|-------|-------|
| Reusable across projects | `skills/` here |
| Specific to one project | `.claude/commands/` inside that project's repo |
| Personal preferences | `~/.claude/` (not tracked here) |

## Bumping a skill version

Just edit the `version:` field in the skill's frontmatter and ship a new release. `duck-ai registry` will surface drift between installed and source versions per agent.
