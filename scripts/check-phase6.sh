#!/bin/sh
set -eu

check-phase5

test -s scripts/package-darwin-arm64.sh
test -s packaging/app/Info.plist.in
test -s packaging/homebrew/Casks/portnado.rb
test -s packaging/homebrew/update-cask.sh
test -s packaging/homebrew/README.md
test -s .github/workflows/release.yml

sh -n scripts/package-darwin-arm64.sh
sh -n packaging/homebrew/update-cask.sh

grep -q 'Portnado.app' packaging/homebrew/Casks/portnado.rb
grep -q 'Contents/Resources/bin/portnado' packaging/homebrew/Casks/portnado.rb
grep -q 'zap trash' packaging/homebrew/Casks/portnado.rb
grep -q 'unidentified-developer' packaging/homebrew/Casks/portnado.rb
grep -q 'dev.portnado.daemon' packaging/homebrew/Casks/portnado.rb

tmp_cask="/tmp/portnado.rb"
cp packaging/homebrew/Casks/portnado.rb "$tmp_cask"
packaging/homebrew/update-cask.sh 0.1.0 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef "$tmp_cask"
grep -q 'version "0.1.0"' "$tmp_cask"
grep -q 'sha256 "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"' "$tmp_cask"

echo "Phase 6 checks passed"
