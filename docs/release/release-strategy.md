# Release Strategy

## Versioning

Use semantic versioning. Initial development target: `0.1.0`.

## Release Artifact

Release archives use:

```text
Portnado-vX.Y.Z-darwin-arm64.zip
```

The archive contains:

```text
Portnado.app
```

The app bundle embeds:

```text
Portnado.app/Contents/Resources/bin/portnado
Portnado.app/Contents/Resources/bin/portnado-daemon
```

## Build Steps

On a version tag:

1. Validate clean source state.
2. Run tests and checks.
3. Build Go CLI for `darwin/arm64`.
4. Build Go daemon for `darwin/arm64`.
5. Build SwiftUI app for Apple Silicon.
6. Embed Go binaries.
7. Verify app bundle structure.
8. Apply ad-hoc signing only if technically required.
9. Create ZIP archive.
10. Calculate SHA-256 checksum.
11. Generate SBOM if practical.
12. Publish GitHub release assets after approval.
13. Generate or update the custom Homebrew Cask.
14. Verify the local artifact with `scripts/verify-release-artifact.sh`.
15. Publish the updated Cask to the Homebrew tap.

The local packaging entry point is:

```bash
PORTNADO_VERSION=0.1.0 make package-darwin-arm64
```

It creates:

```text
dist/Portnado-v0.1.0-darwin-arm64.zip
dist/Portnado-v0.1.0-darwin-arm64.zip.sha256
```

## Signing

The project currently has no Developer ID certificate and no notarization. Documentation must not claim Apple has verified the developer. If ad-hoc signing is used, it is only for local bundle integrity and does not provide Apple trust.

## Homebrew Cask

The cask lives at:

```text
packaging/homebrew/Casks/portnado.rb
```

It installs `Portnado.app`, exposes the embedded CLI as `portnado`, declares Apple Silicon and macOS requirements, includes uninstall and `zap`, and explains unsigned/unnotarized caveats. It must not run Gatekeeper bypass commands.

## Tap Automation

The tap repository is `marcel-breuer/homebrew-tap`, which Homebrew exposes
as `marcel-breuer/tap`. Users install with:

```bash
brew install --cask marcel-breuer/tap/portnado
```

Initial bootstrapping must not require credentials. Tag releases publish to the
tap only when `HOMEBREW_TAP_TOKEN` is configured with contents write access to
`marcel-breuer/homebrew-tap`.

Use:

```bash
packaging/homebrew/update-cask.sh 0.1.0 "$(cat dist/Portnado-v0.1.0-darwin-arm64.zip.sha256)"
packaging/homebrew/publish-tap.sh 0.1.0 "$(cat dist/Portnado-v0.1.0-darwin-arm64.zip.sha256)"
```

Publishing to the tap must not require broader account permissions.
