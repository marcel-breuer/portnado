#!/bin/sh
set -eu

if [ "$(uname -s)" != "Darwin" ]; then
  echo "Packaging Portnado.app requires macOS host tooling." >&2
  exit 1
fi

version="${PORTNADO_VERSION:-0.1.0-dev}"
build="${PORTNADO_BUILD:-1}"
root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
dist_dir="$root/dist"
build_dir="$root/build/release/darwin-arm64"
app_dir="$dist_dir/Portnado.app"
archive_name="Portnado-v${version}-darwin-arm64.zip"
archive_path="$dist_dir/$archive_name"
ldflags="-s -w -X github.com/marcel-breuer/portnado/internal/version.Version=$version"

rm -rf "$build_dir" "$app_dir" "$archive_path" "$archive_path.sha256"
mkdir -p "$build_dir" "$app_dir/Contents/MacOS" "$app_dir/Contents/Resources/bin"

echo "Building Go CLI and daemon for darwin/arm64"
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "$ldflags" -o "$app_dir/Contents/Resources/bin/portnado" ./cmd/portnado
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "$ldflags" -o "$app_dir/Contents/Resources/bin/portnado-daemon" ./cmd/portnado-daemon

echo "Building Swift menu bar executable"
swift build -c release --arch arm64 --package-path "$root/apps/Portnado"
swift_binary="$root/apps/Portnado/.build/arm64-apple-macosx/release/Portnado"
if [ ! -x "$swift_binary" ]; then
  swift_binary="$(find "$root/apps/Portnado/.build" -path '*/release/Portnado' -type f -perm -111 | head -n 1)"
fi
if [ ! -x "$swift_binary" ]; then
  echo "Could not find built Swift executable." >&2
  exit 1
fi
cp "$swift_binary" "$app_dir/Contents/MacOS/Portnado"

sed \
  -e "s/{{VERSION}}/$version/g" \
  -e "s/{{BUILD}}/$build/g" \
  "$root/packaging/app/Info.plist.in" > "$app_dir/Contents/Info.plist"

chmod 755 "$app_dir/Contents/MacOS/Portnado" "$app_dir/Contents/Resources/bin/portnado" "$app_dir/Contents/Resources/bin/portnado-daemon"

if [ "${PORTNADO_CODESIGN:-skip}" = "adhoc" ]; then
  echo "Applying ad-hoc signature"
  codesign --force --deep --sign - "$app_dir"
fi

echo "Creating release archive"
mkdir -p "$dist_dir"
if command -v ditto >/dev/null 2>&1; then
  (cd "$dist_dir" && COPYFILE_DISABLE=1 ditto -c -k --norsrc --keepParent "Portnado.app" "$archive_name")
else
  (cd "$dist_dir" && COPYFILE_DISABLE=1 zip -qr "$archive_name" "Portnado.app")
fi

shasum -a 256 "$archive_path" | awk '{print $1}' > "$archive_path.sha256"
echo "Archive: $archive_path"
echo "SHA-256: $(cat "$archive_path.sha256")"
