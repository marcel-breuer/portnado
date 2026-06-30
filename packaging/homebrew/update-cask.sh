#!/bin/sh
set -eu

usage() {
  echo "usage: update-cask.sh VERSION SHA256 [CASK_PATH]" >&2
  exit 2
}

version="${1:-}"
sha256="${2:-}"
cask_path="${3:-$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/Casks/portnado.rb}"

[ -n "$version" ] || usage
[ -n "$sha256" ] || usage
[ -f "$cask_path" ] || {
  echo "cask file not found: $cask_path" >&2
  exit 1
}

case "$sha256" in
  *[!0123456789abcdef]*)
    echo "sha256 must be lowercase hexadecimal" >&2
    exit 2
    ;;
esac

if [ "${#sha256}" -ne 64 ]; then
  echo "sha256 must be 64 characters" >&2
  exit 2
fi

tmp="$(mktemp)"
sed \
  -e "s/^  version \".*\"/  version \"$version\"/" \
  -e "s/^  sha256 \".*\"/  sha256 \"$sha256\"/" \
  "$cask_path" > "$tmp"
mv "$tmp" "$cask_path"

echo "Updated $cask_path to $version"
