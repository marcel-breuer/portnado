#!/bin/sh
set -eu

check-phase5

test -s scripts/package-darwin-arm64.sh
test -s packaging/app/Info.plist.in
test -s packaging/homebrew/Casks/portnado.rb
test -s packaging/homebrew/publish-tap.sh
test -s packaging/homebrew/update-cask.sh
test -s packaging/homebrew/README.md
test -s .github/workflows/release.yml

sh -n scripts/package-darwin-arm64.sh
sh -n packaging/homebrew/publish-tap.sh
sh -n packaging/homebrew/update-cask.sh

grep -q 'Portnado.app' packaging/homebrew/Casks/portnado.rb
grep -q 'Contents/Resources/bin/portnado' packaging/homebrew/Casks/portnado.rb
grep -q 'zap trash' packaging/homebrew/Casks/portnado.rb
grep -q 'unidentified-developer' packaging/homebrew/Casks/portnado.rb
grep -q 'malicious software' packaging/homebrew/Casks/portnado.rb
grep -q 'dev.portnado.daemon' packaging/homebrew/Casks/portnado.rb
grep -q 'marcel-breuer/homebrew-tap' packaging/homebrew/publish-tap.sh
grep -q 'marcel-breuer/tap/portnado' README.md
grep -q 'marcel-breuer/tap/portnado' docs/installation.md
grep -q 'HOMEBREW_TAP_TOKEN' .github/workflows/release.yml
grep -q 'publish-tap.sh' .github/workflows/release.yml

tmp_cask="/tmp/portnado.rb"
cp packaging/homebrew/Casks/portnado.rb "$tmp_cask"
packaging/homebrew/update-cask.sh 0.1.0 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef "$tmp_cask"
grep -q 'version "0.1.0"' "$tmp_cask"
grep -q 'sha256 "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"' "$tmp_cask"

tmp_tap="$(mktemp -d)"
trap 'rm -rf "$tmp_tap"' EXIT
git init -b main "$tmp_tap" >/dev/null
HOMEBREW_TAP_DIR="$tmp_tap" HOMEBREW_TAP_PUSH=0 packaging/homebrew/publish-tap.sh 0.1.0 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef >/tmp/portnado-publish-tap.log
grep -q 'version "0.1.0"' "$tmp_tap/Casks/portnado.rb"
grep -q 'sha256 "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"' "$tmp_tap/Casks/portnado.rb"
git -C "$tmp_tap" log --oneline -1 | grep -q 'chore: update portnado cask to v0.1.0'

echo "Phase 6 checks passed"
