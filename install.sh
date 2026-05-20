#!/usr/bin/env bash
set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILLS_TARGET="${HOME}/.claude-work/skills"
CLAUDE_DIR="${HOME}/.claude"

usage() {
  echo "Usage: ./install.sh [--skills] [--commands] [--all]"
  echo ""
  echo "  --skills    Symlink skills into ~/.claude-work/skills/"
  echo "  --commands  Symlink commands into ~/.claude/commands/"
  echo "  --all       Install everything (default)"
  exit 0
}

install_skills() {
  echo "→ Installing skills..."
  mkdir -p "$SKILLS_TARGET"
  for skill_dir in "$REPO_DIR/skills"/*/; do
    skill_name=$(basename "$skill_dir")
    target="$SKILLS_TARGET/$skill_name"
    if [ -L "$target" ]; then
      echo "  ↻ $skill_name (already linked, updating)"
      rm "$target"
    elif [ -d "$target" ]; then
      echo "  ⚠ $skill_name exists as directory — skipping (remove manually to replace)"
      continue
    fi
    ln -s "$skill_dir" "$target"
    echo "  ✓ $skill_name"
  done
}

install_commands() {
  echo "→ Installing commands..."
  mkdir -p "$CLAUDE_DIR/commands"
  for cmd_file in "$REPO_DIR/claude/commands"/*.md; do
    [ -f "$cmd_file" ] || continue
    cmd_name=$(basename "$cmd_file")
    target="$CLAUDE_DIR/commands/$cmd_name"
    if [ -L "$target" ]; then
      rm "$target"
    fi
    ln -s "$cmd_file" "$target"
    echo "  ✓ $cmd_name"
  done
}

# Parse args
if [ $# -eq 0 ] || [ "$1" = "--all" ]; then
  install_skills
  install_commands
else
  for arg in "$@"; do
    case $arg in
      --skills)   install_skills ;;
      --commands) install_commands ;;
      --help|-h)  usage ;;
      *)          echo "Unknown option: $arg"; usage ;;
    esac
  done
fi

echo ""
echo "Done. Run ./doctor.sh to verify."
