#!/bin/bash
# Generate platform-specific AI rules from docs/AGENT_RULES.md
set -e

SRC="docs/AGENT_RULES.md"

if [ ! -f "$SRC" ]; then
  echo "Error: $SRC not found" >&2
  exit 1
fi

# Extract a section by name: everything between ## [Name] and the next ## [ or EOF
extract() {
  local name="$1"
  awk -v name="$name" '
    $0 == "## [" name "]" { found=1; next }
    /^## \[/ { found=0 }
    found { print }
  ' "$SRC"
}

SECTIONS=(
  "Project Overview"
  "Zero External Dependencies"
  "doc.go Facade"
  "Dependency Direction"
  "Naming"
  "Testing"
  "Code Style"
  "Documentation"
  "Skill"
  "Impact Analysis"
  "Common Pitfalls"
  "Auto-Config Priority"
  "Verification"
)

generate_file() {
  local file="$1" header="$2"
  mkdir -p "$(dirname "$file")"
  {
    echo "$header"
    echo ""
    for section in "${SECTIONS[@]}"; do
      content=$(extract "$section")
      if [ -n "$content" ]; then
        echo "## $section"
        echo ""
        echo "$content"
        echo ""
      fi
    done
  } > "$file"
}

# Check mode: generate to temp dir and diff
if [ "${1:-}" = "check" ]; then
  TMPDIR=$(mktemp -d)
  trap 'rm -rf "$TMPDIR"' EXIT

  generate_file "$TMPDIR/.cursorrules" "# Cursor Rules for enhance"
  generate_file "$TMPDIR/.windsurfrules" "# Windsurf Rules for enhance"
  generate_file "$TMPDIR/.codex" "# Codex Configuration for enhance"
  generate_file "$TMPDIR/.github/copilot-instructions.md" "# GitHub Copilot Instructions for enhance"

  ok=true
  for f in .cursorrules .windsurfrules .codex .github/copilot-instructions.md; do
    if ! diff -q "$TMPDIR/$f" "$f" >/dev/null 2>&1; then
      echo "DIFF: $f"
      diff --color -u "$f" "$TMPDIR/$f" || true
      ok=false
    fi
  done
  if [ "$ok" = true ]; then
    echo "All platform files match AGENT_RULES.md"
  else
    echo "Platform files are out of sync. Run: make skill-sync"
    exit 1
  fi
  exit 0
fi

# Generate mode
generate_file ".cursorrules" "# Cursor Rules for enhance"
generate_file ".windsurfrules" "# Windsurf Rules for enhance"
generate_file ".codex" "# Codex Configuration for enhance"
generate_file ".github/copilot-instructions.md" "# GitHub Copilot Instructions for enhance"

echo "Generated: .cursorrules .windsurfrules .codex .github/copilot-instructions.md"