#!/bin/sh
set -eu

usage() {
  echo "usage: publish-tap.sh VERSION SHA256" >&2
  exit 2
}

version="${1:-}"
sha256="${2:-}"

[ -n "$version" ] || usage
[ -n "$sha256" ] || usage

case "$version" in
  *[!/0-9A-Za-z._-]* | */* | -*)
    echo "version contains unsupported characters" >&2
    exit 2
    ;;
esac

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

root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
source_cask="$root/packaging/homebrew/Casks/portnado.rb"
tap_repo="${HOMEBREW_TAP_REPO:-marcel-breuer/homebrew-portnado}"
tap_branch="${HOMEBREW_TAP_BRANCH:-main}"
tap_dir="${HOMEBREW_TAP_DIR:-}"
push="${HOMEBREW_TAP_PUSH:-1}"
token="${HOMEBREW_TAP_TOKEN:-}"
tmp_dir=""

cleanup() {
  if [ -n "$tmp_dir" ]; then
    rm -rf "$tmp_dir"
  fi
}
trap cleanup EXIT

[ -f "$source_cask" ] || {
  echo "cask file not found: $source_cask" >&2
  exit 1
}

case "$tap_repo" in
  *[!0-9A-Za-z._/-]* | /* | */../* | ../*)
    echo "tap repository must be an owner/repository name" >&2
    exit 2
    ;;
esac

if [ "$push" != "0" ] && [ -z "$token" ]; then
  echo "HOMEBREW_TAP_TOKEN is required when HOMEBREW_TAP_PUSH is enabled" >&2
  exit 2
fi

if [ -z "$tap_dir" ]; then
  tmp_dir="$(mktemp -d)"
  tap_dir="$tmp_dir/tap"
  clone_url="https://github.com/$tap_repo.git"
  if [ -n "$token" ]; then
    clone_url="https://x-access-token:$token@github.com/$tap_repo.git"
  fi
  git clone --depth 1 --branch "$tap_branch" "$clone_url" "$tap_dir"
else
  mkdir -p "$tap_dir"
  if [ ! -d "$tap_dir/.git" ]; then
    git init -b "$tap_branch" "$tap_dir"
  fi
fi

mkdir -p "$tap_dir/Casks"
cp "$source_cask" "$tap_dir/Casks/portnado.rb"
"$root/packaging/homebrew/update-cask.sh" "$version" "$sha256" "$tap_dir/Casks/portnado.rb" >/dev/null

git -C "$tap_dir" config user.name "${HOMEBREW_TAP_GIT_NAME:-Portnado Release}"
git -C "$tap_dir" config user.email "${HOMEBREW_TAP_GIT_EMAIL:-release@portnado.local}"
git -C "$tap_dir" add Casks/portnado.rb

if git -C "$tap_dir" diff --cached --quiet -- Casks/portnado.rb; then
  echo "Homebrew tap already contains Portnado $version"
  exit 0
fi

git -C "$tap_dir" commit -m "chore: update portnado cask to v$version"

if [ "$push" = "0" ]; then
  echo "Prepared Homebrew tap update in $tap_dir"
  exit 0
fi

git -C "$tap_dir" push origin "HEAD:$tap_branch"
echo "Published Homebrew tap update for Portnado $version"
