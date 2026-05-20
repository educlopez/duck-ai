# Adding a skill

1. Create `skills/<skill-name>/SKILL.md`
2. Frontmatter required:
   ```yaml
   ---
   name: skill-name
   description: >
     When to trigger this skill. Be specific — Claude uses this to decide
     whether to invoke the skill. Include contexts, project types, keywords.
   ---
   ```
3. Run `./install.sh --skills` to symlink it
4. Run `./doctor.sh` to verify

## Adding a command

1. Create `claude/commands/<command-name>.md`
2. Available in Claude Code as `/<command-name>`
3. Run `./install.sh --commands` to symlink it

## Skill scope

| Scope | Where |
|-------|-------|
| Reusable across projects | `skills/` here |
| Specific to one project | `.claude/commands/` inside that project's repo |
| Personal preferences | `~/.claude/` (not tracked here) |
