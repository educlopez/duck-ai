#!/usr/bin/env bash

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILLS_TARGET="${HOME}/.claude-work/skills"
CLAUDE_DIR="${HOME}/.claude"
PASS=0
FAIL=0

ok()   { echo "  ✓ $1"; ((PASS++)); }
fail() { echo "  ✗ $1"; ((FAIL++)); }
section() { echo ""; echo "── $1"; }

section "Claude Code"
if command -v claude &>/dev/null; then
  ok "claude CLI installed ($(claude --version 2>/dev/null | head -1))"
else
  fail "claude CLI not found — install from https://claude.ai/code"
fi

section "Node / pnpm"
if command -v node &>/dev/null; then
  ok "node $(node --version)"
else
  fail "node not found"
fi

if command -v pnpm &>/dev/null; then
  pnpm_ver=$(pnpm --version)
  major=$(echo "$pnpm_ver" | cut -d. -f1)
  if [ "$major" -ge 11 ]; then
    ok "pnpm $pnpm_ver"
  else
    fail "pnpm $pnpm_ver — upgrade to pnpm 11+ (npm install -g pnpm@latest)"
  fi
else
  fail "pnpm not found — npm install -g pnpm"
fi

section "PHP / Composer"
if command -v composer &>/dev/null; then
  ok "composer $(composer --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')"
else
  ok "composer not found (OK if no PHP projects)"
fi

section "Skills"
for skill_dir in "$REPO_DIR/skills"/*/; do
  skill_name=$(basename "$skill_dir")
  target="$SKILLS_TARGET/$skill_name"
  if [ -L "$target" ] && [ -d "$target" ]; then
    ok "skill: $skill_name (linked)"
  elif [ -d "$target" ]; then
    fail "skill: $skill_name (exists but not a symlink — run ./install.sh --skills)"
  else
    fail "skill: $skill_name (not installed — run ./install.sh --skills)"
  fi
done

section "Commands"
cmd_count=0
for cmd_file in "$REPO_DIR/claude/commands"/*.md; do
  [ -f "$cmd_file" ] || continue
  cmd_name=$(basename "$cmd_file")
  target="$CLAUDE_DIR/commands/$cmd_name"
  if [ -L "$target" ]; then
    ok "command: $cmd_name (linked)"
  else
    fail "command: $cmd_name (not installed — run ./install.sh --commands)"
  fi
  ((cmd_count++))
done
[ "$cmd_count" -eq 0 ] && echo "  (no commands defined yet)"

echo ""
echo "────────────────────────────"
echo "  $PASS passed  |  $FAIL failed"
echo "────────────────────────────"
[ "$FAIL" -gt 0 ] && exit 1 || exit 0
