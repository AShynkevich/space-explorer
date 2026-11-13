#!/usr/bin/env bash
# Build Space Explorer (GUI app), space-explorer-tui (terminal UI) and scantool
# (dev CLI) and collect the results in build/bin/<os>-<arch>/.
#
# Usage: scripts/build.sh [target ...]
#   target = os/arch, e.g. darwin/arm64 windows/amd64 linux/amd64
#   no target = this machine, plus windows/amd64 when building on macOS/Linux
#
# Notes: macOS and Windows GUIs cross-build from macOS. The Linux GUI needs GTK3 +
# WebKitGTK, so on a non-Linux host it is built inside Docker (scripts/linux-builder.Dockerfile);
# without Docker it is skipped with a warning. The CLI tools always cross-compile.
# A non-native Linux arch runs under emulation and is slow the first time.
set -euo pipefail

# Ctrl+C must stop the whole script, not just the running child (e.g. docker).
trap 'echo; echo "interrupted" >&2; exit 130' INT TERM

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/build/bin"
cd "$ROOT"

# Toolchain dirs are not on PATH by default on the dev machine; harmless elsewhere.
for d in "$HOME/DEV/GO/go/bin" "$HOME/DEV/GO/workspace/bin" "$HOME/.rd/bin"; do
  [ -d "$d" ] && PATH="$d:$PATH"
done
export PATH

need() { command -v "$1" >/dev/null || { echo "error: '$1' not found in PATH" >&2; exit 1; }; }
need go; need npm; need wails

host_os="$(go env GOOS)"
host_arch="$(go env GOARCH)"

if [ "$#" -gt 0 ]; then
  targets=("$@")
else
  targets=(darwin/arm64 darwin/amd64 windows/amd64 windows/arm64 linux/amd64 linux/arm64)
fi

docker_ready() { command -v docker >/dev/null && docker info >/dev/null 2>&1; }

# Can this host build the Wails GUI for the given OS?
gui_supported() {
  case "$1" in
    darwin) [ "$host_os" = darwin ] ;;
    windows) return 0 ;;          # cross-compiles without cgo
    linux) [ "$host_os" = linux ] || docker_ready ;;
    *) return 1 ;;
  esac
}

gui_name() { [ "$1" = windows ] && echo space-explorer.exe || echo space-explorer; }

stage_gui() { # os dest
  local os="$1" dest="$2" built
  case "$os" in
    darwin) built="$OUT/space-explorer.app" ;;
    *) built="$OUT/$(gui_name "$os")" ;;
  esac
  rm -rf "$dest/$(basename "$built")"
  mv "$built" "$dest/"
}

# Linux GUI from a non-Linux host: compile inside a Linux container.
# The frontend is built here first (skipped in the container with -s).
# Every step is chained with `|| return 1`: inside `if build_gui ...` set -e is off,
# so without it a failed step would be ignored and the next one would still run.
build_gui_docker() { # arch
  local arch="$1" image="space-explorer-linux-builder:$1"
  if ! docker image inspect "$image" >/dev/null 2>&1; then
    echo "    building Docker image $image (first time only; slow under emulation, several minutes)..."
    docker build --progress=plain --platform "linux/$arch" -t "$image" \
      -f scripts/linux-builder.Dockerfile scripts 2>&1 | sed 's/^/      /' ; [ "${PIPESTATUS[0]}" -eq 0 ] || return 1
  fi
  (cd frontend && npm install --silent --no-audit --no-fund && npm run build --silent) >/dev/null || return 1
  touch frontend/dist/gitkeep
  echo "    compiling the GUI in the container..."
  docker run --rm --platform "linux/$arch" -v "$ROOT":/src -w /src \
    -v space-explorer-gomod:/go/pkg/mod -v space-explorer-gocache:/root/.cache/go-build \
    "$image" wails build -s -clean=false -tags webkit2_41 -platform "linux/$arch" -o space-explorer >/dev/null || return 1
}

build_gui() { # os arch dest
  local os="$1" arch="$2" dest="$3"
  # The frontend build deletes the placeholder that go:embed needs.
  mkdir -p frontend/dist && touch frontend/dist/gitkeep
  # Wails always writes into build/bin; clear leftovers so we move the right thing.
  rm -rf "$OUT/space-explorer.app" "$OUT/$(gui_name "$os")"
  if [ "$os" = linux ] && [ "$host_os" != linux ]; then
    build_gui_docker "$arch" || return 1
  else
    wails build -clean=false -platform "$os/$arch" -o "$(gui_name "$os")" >/dev/null || return 1
  fi
  touch frontend/dist/gitkeep
  stage_gui "$os" "$dest" || return 1
}

build_cli() { # os arch dest cmd-name
  local os="$1" arch="$2" dest="$3" name="$4" file="$4"
  [ "$os" = windows ] && file="$name.exe"
  GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "-s -w" -o "$dest/$file" "./cmd/$name"
}

failed=0
for t in "${targets[@]}"; do
  os="${t%/*}"; arch="${t#*/}"
  if [ "$os" = "$t" ] || [ -z "$arch" ]; then
    echo "error: bad target '$t' (expected os/arch)" >&2; exit 1
  fi
  dest="$OUT/$os-$arch"
  mkdir -p "$dest"
  echo "==> $os/$arch -> ${dest#"$ROOT"/}"

  for tool in space-explorer-tui scantool; do
    if build_cli "$os" "$arch" "$dest" "$tool"; then echo "    $tool ok"
    else echo "    $tool FAILED" >&2; failed=1; fi
  done

  if gui_supported "$os"; then
    if build_gui "$os" "$arch" "$dest"; then echo "    space-explorer ok"
    else echo "    space-explorer FAILED" >&2; failed=1; fi
  else
    echo "    space-explorer skipped: needs a $os host or Docker (see header notes)" >&2
  fi
done

echo
find "$OUT" -mindepth 2 -maxdepth 2 | sed "s|$ROOT/||" | sort
exit "$failed"
