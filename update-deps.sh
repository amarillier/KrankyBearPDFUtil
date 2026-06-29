#!/usr/bin/env bash
# Update Go dependencies across every module in this repo (root + sub-modules) and tidy,
# then confirm everything still builds. Handy for clearing Dependabot alerts.
#
# Usage:
#   ./update-deps.sh                       # bump the common security-flagged x/* packages (those present)
#   ./update-deps.sh golang.org/x/net      # bump only the package(s) you name (to @latest)
#   ./update-deps.sh golang.org/x/net@v0.38.0   # pin an exact version (e.g. to match a Dependabot PR)
#   ./update-deps.sh -u                    # blanket: go get -u ./...  (latest minor/patch of ALL deps)
#   ./update-deps.sh -p                    # blanket: go get -u=patch ./... (patch releases only — safer)
#
# Targeted (default / named packages) is usually the right choice for Dependabot: minimal
# blast radius, easy to confirm the build still passes.

set -uo pipefail

usage() {
  cat <<'EOF'
update-deps.sh - update Go deps across all modules in this repo, tidy, and build-check.

Usage:
  ./update-deps.sh                          bump common security-flagged x/* packages (those present)
  ./update-deps.sh golang.org/x/net         bump only the named package(s) to @latest
  ./update-deps.sh golang.org/x/net@v0.38.0 pin an exact version (e.g. match a Dependabot PR)
  ./update-deps.sh -u                       blanket: go get -u ./...      (latest of ALL deps)
  ./update-deps.sh -p                       blanket: go get -u=patch ./... (patch releases only)
  ./update-deps.sh -h                       this help

Targeted (default / named packages) is usually best for Dependabot: minimal blast radius.
EOF
}

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT" || exit 1

# Default set: the golang.org/x/* modules that most commonly get security advisories.
# Only those actually present in a given module are touched.
DEFAULT_PKGS=(
  golang.org/x/crypto
  golang.org/x/image
  golang.org/x/net
  golang.org/x/sys
  golang.org/x/text
  golang.org/x/term
)

MODE="targeted"   # targeted | update | patch
PKGS=()
for arg in "$@"; do
  case "$arg" in
    -u|--update) MODE="update" ;;
    -p|--patch)  MODE="patch" ;;
    -h|--help)   usage; exit 0 ;;
    -*)          echo "Unknown option: $arg" >&2; exit 1 ;;
    *)           PKGS+=("$arg") ;;
  esac
done
[[ ${#PKGS[@]} -eq 0 ]] && PKGS=("${DEFAULT_PKGS[@]}")

# Find every module (skip vendor/.git). macOS bash 3.2-safe (no mapfile).
MODULES=()
while IFS= read -r m; do MODULES+=("$(dirname "$m")"); done \
  < <(find . -name go.mod -not -path './.git/*' -not -path '*/vendor/*' | sort)

if [[ ${#MODULES[@]} -eq 0 ]]; then
  echo "No go.mod found under $ROOT" >&2
  exit 1
fi

echo "Modules: ${MODULES[*]}"
echo "Mode: $MODE"
[[ "$MODE" == "targeted" ]] && echo "Packages: ${PKGS[*]}"
echo ""

FAILED=()
for d in "${MODULES[@]}"; do
  echo "════════ $d ════════"
  (
    cd "$d" || exit 1
    export GOFLAGS=-mod=mod
    case "$MODE" in
      update) go get -u ./... ;;
      patch)  go get -u=patch ./... ;;
      targeted)
        for p in "${PKGS[@]}"; do
          base="${p%@*}"   # strip any @version for the presence check
          if go list -m "$base" >/dev/null 2>&1; then
            if go get "$p" 2>/dev/null; then
              echo "  ✓ $p"
            else
              echo "  ! could not bump $p"
            fi
          fi
        done
        ;;
    esac
    echo "  → go mod tidy"
    go mod tidy
  ) || { echo "  ✗ dependency update failed in $d"; FAILED+=("$d (deps)"); }
done

echo ""
echo "════════ build check (CGO enabled) ════════"
for d in "${MODULES[@]}"; do
  if ( cd "$d" && CGO_ENABLED=1 go build ./... ); then
    echo "  ✓ $d builds"
  else
    echo "  ✗ $d FAILED to build"
    FAILED+=("$d (build)")
  fi
done

echo ""
if [[ ${#FAILED[@]} -eq 0 ]]; then
  echo "✓ All modules updated, tidied, and building."
  echo "  Review: git diff -- '*go.mod' '*go.sum'   then commit & push."
else
  echo "⚠ Issues in: ${FAILED[*]}"
  echo "  Inspect the output above before committing."
  exit 1
fi

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
