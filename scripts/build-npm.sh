#!/usr/bin/env bash
# Cross-compile dahmer for every supported (OS, arch) pair and stage each
# binary into its npm sub-package under npm/. Run from the repo root.
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${VERSION:-$(node -p "require('./npm/dahmer/package.json').version")}"
echo "Building dahmer v$VERSION"

# (goos, goarch, npm-platform, npm-arch, bin-name)
TARGETS=(
  "darwin  arm64 darwin arm64 dahmer"
  "darwin  amd64 darwin x64   dahmer"
  "linux   arm64 linux  arm64 dahmer"
  "linux   amd64 linux  x64   dahmer"
  "windows arm64 win32  arm64 dahmer.exe"
  "windows amd64 win32  x64   dahmer.exe"
)

for row in "${TARGETS[@]}"; do
  # shellcheck disable=SC2086
  set -- $row
  GOOS=$1 GOARCH=$2 NPM_OS=$3 NPM_ARCH=$4 BIN=$5

  pkg="npm/dahmer-${NPM_OS}-${NPM_ARCH}"
  mkdir -p "$pkg/bin"

  echo "  -> $pkg/bin/$BIN  (GOOS=$GOOS GOARCH=$GOARCH)"
  GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o "$pkg/bin/$BIN" .

  cat > "$pkg/package.json" <<JSON
{
  "name": "dahmer-${NPM_OS}-${NPM_ARCH}",
  "version": "${VERSION}",
  "description": "dahmer prebuilt binary for ${NPM_OS}-${NPM_ARCH}",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/ASHUTOSH-SWAIN-GIT/dahmer.git"
  },
  "homepage": "https://github.com/ASHUTOSH-SWAIN-GIT/dahmer",
  "os": ["${NPM_OS}"],
  "cpu": ["${NPM_ARCH}"],
  "files": ["bin/${BIN}"]
}
JSON
done

echo "Done. Publish order:"
echo "  1) cd npm/dahmer-*/ && npm publish --access public   (each platform pkg)"
echo "  2) cd npm/dahmer    && npm publish --access public   (main pkg last)"
