---
name: duck-ai-release
description: >
  Workflow for adding skills/commands to duck-ai and shipping a new release.
  Use when the user asks how to release duck-ai, ship a new duck-ai version, add a
  skill or command to duck-ai, bump a duck-ai version, "publicar duck-ai", "subir
  duck-ai", "nueva versión de duck-ai", or anything involving editing this repo
  and getting the change to the Cinetic team's machines.
version: "0.1.0"
---

# duck-ai release workflow

Reference for adding/updating skills, commands, and shipping releases for [duck-ai](https://github.com/educlopez/duck-ai). The tool lives at `/Users/eduardocalvolopez/Developer/Personal/duck-ai` (on Eduardo's machine).

## When to use this skill

- "How do I add a skill to duck-ai?"
- "Release a new version of duck-ai" / "ship duck-ai" / "tag duck-ai vX.Y.Z"
- "Update duck-ai" / "bump the duck-ai version"
- Anything that involves editing files under `skills/`, `claude/commands/`, or shipping the change to other team members.

## TL;DR

```bash
# 1. Edit the skill/command file
# 2. Bump its `version:` in frontmatter (semver, conservative)
# 3. Test locally
go build -o duck-ai .
./duck-ai update                    # re-link with backup
./duck-ai registry                  # confirm new version + ok status

# 4. Commit using Conventional Commits, in English
git add skills/<name>/ claude/commands/<name>.md
git commit -m "feat(skills): add <name>"   # or fix:, refactor:, docs:, chore:

# 5. Push
git push

# 6. Tag a release (this is what propagates to the team)
git tag vX.Y.Z
git push origin vX.Y.Z

# GitHub Actions runs goreleaser → publishes binaries + auto-changelog
```

Team members pick it up with:

```bash
curl -fsSL https://raw.githubusercontent.com/educlopez/duck-ai/main/install.sh | bash
duck-ai update
```

## Adding a new skill

1. Create `skills/<skill-name>/SKILL.md` with frontmatter:
   ```yaml
   ---
   name: <skill-name>
   description: >
     When to trigger this skill. Be specific — Claude uses this to decide
     whether to invoke. Include keywords the user might say, project types,
     and contexts.
   version: "0.1.0"
   ---
   ```
2. Write the skill body in markdown. Sections, code blocks, tables — whatever helps Claude execute the skill. Reference existing skills in this repo for style.
3. Locally: `go build -o duck-ai . && ./duck-ai update` — the symlink will appear in `~/.claude/skills/<skill-name>/`.
4. Validate: `./duck-ai doctor` (managed count goes up) and `./duck-ai registry` (entry shows `v0.1.0 ok`).
5. Skip to **Shipping a release** below.

## Adding a new command

1. Create `claude/commands/<command-name>.md` with frontmatter:
   ```yaml
   ---
   name: <command-name>
   description: One-line summary of what the command does.
   version: "0.1.0"
   ---
   ```
2. Body: instructions for the slash command. Available in Claude Code as `/<command-name>` after install.
3. Locally: `./duck-ai update`.
4. Skip to **Shipping a release** below.

## Bumping an existing skill or command

1. Edit the file.
2. Bump `version:` in the frontmatter. Semver guidance:
   - Patch (0.1.0 → 0.1.1): typo fixes, doc clarifications, no behavior change.
   - Minor (0.1.0 → 0.2.0): new functionality inside the skill, backward-compatible.
   - Major (0.1.0 → 1.0.0): breaking change in expected user-facing behavior (rare for a skill).
3. Locally: `./duck-ai update && ./duck-ai registry` — confirm new version shows `ok`, old `drift` if you also have an older symlink elsewhere.

## Shipping a release

```bash
git add <files>
git commit -m "feat(skills): add <name>"
# Conventional Commit prefixes: feat, fix, refactor, docs, chore, ci, test
git push
git tag vX.Y.Z
git push origin vX.Y.Z
```

Version the **duck-ai release**, not the skill: skills carry their own versions inside frontmatter, the tag versions the binary + the set of skills shipped together.

GitHub Actions (`.github/workflows/release.yml`) runs goreleaser, which:
- Builds for darwin/linux/windows × amd64/arm64
- Uploads tarballs + checksums to the GitHub release
- Writes a release changelog grouped by Features / Bug Fixes / Refactor / Others (filters out `docs:`, `test:`, `ci:`, `chore:`)

If the changelog auto-text needs richer notes (highlights, breaking changes, etc.), edit the release after publish:

```bash
gh release edit vX.Y.Z --notes "$(cat <<'EOF'
## Highlights
- ...

## Skills
- new: ...
- updated: ...
EOF
)"
```

## Team upgrade path

```bash
curl -fsSL https://raw.githubusercontent.com/educlopez/duck-ai/main/install.sh | bash
duck-ai update     # re-symlinks skills + commands, backs up any conflicts
duck-ai doctor     # verify
duck-ai registry   # see what's installed per agent + versions
```

Windows: download the latest zip from https://github.com/educlopez/duck-ai/releases and replace `duck-ai.exe` on the PATH.

## Gotchas

- **NEVER skip the `version:` field** on skill/command frontmatter. Without it, `duck-ai registry` reports `unversioned`.
- **Conventional Commits are mandatory** — they drive the changelog grouping. `chore:` commits get filtered out of release notes.
- **`duck-ai update` makes backups** before replacing files. If you blow something away, `duck-ai update --list-backups` and `--restore <ts>` will recover it.
- **The first run on a new agent dir** (Claude / Codex / OpenCode / generic) appears as `missing → installed` for every entry. That's expected, not drift.
- **Skills only ship via tags.** Pushing to `main` without tagging publishes nothing — GH Actions release workflow only fires on `v*` tags.

## Useful commands

| Command | What |
|---------|------|
| `./duck-ai` | Interactive TUI (Install / Update / Doctor / Registry / Quit) |
| `./duck-ai update --dry-run` | Show what update would change without touching disk |
| `./duck-ai update --restore <ts>` | Recover from a backup batch |
| `./duck-ai registry --source` | List what the repo ships (without scanning installed) |
| `./duck-ai registry --all` | Include foreign / unmanaged entries from other tooling |
| `./duck-ai version` | Print version (release tag injected at build time) |
