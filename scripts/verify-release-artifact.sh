#!/bin/sh
set -eu

version="${1:-0.1.0}"
root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
archive="$root/dist/Portnado-v${version}-darwin-arm64.zip"
checksum_file="$archive.sha256"
cask="$root/packaging/homebrew/Casks/portnado.rb"

test -f "$archive"
test -f "$checksum_file"
test -f "$cask"

actual="$(shasum -a 256 "$archive" | awk '{print $1}')"
expected="$(cat "$checksum_file")"
if [ "$actual" != "$expected" ]; then
  echo "checksum mismatch: $actual != $expected" >&2
  exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
unzip -q "$archive" -d "$tmp"

app="$tmp/Portnado.app"
test -x "$app/Contents/MacOS/Portnado"
test -x "$app/Contents/Resources/bin/portnado"
test -x "$app/Contents/Resources/bin/portnado-daemon"
plutil -lint "$app/Contents/Info.plist" >/dev/null

if find "$tmp" -name '._*' | grep -q .; then
  echo "archive contains AppleDouble metadata" >&2
  exit 1
fi

"$app/Contents/Resources/bin/portnado" --version | grep -q "portnado $version"
grep -q "sha256 \"$actual\"" "$cask"
grep -q "version \"$version\"" "$cask"

echo "Release artifact verified: $archive"
