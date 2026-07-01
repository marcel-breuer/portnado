# Homebrew Tap

The tap repository is:

```text
marcel-breuer/homebrew-tap
```

Homebrew users install from the tap name `marcel-breuer/tap`:

```bash
brew install --cask marcel-breuer/tap/portnado
```

The tap must contain `Casks/portnado.rb`. Homebrew loads casks from a top-level
`Casks` directory in a tap.

Update the cask locally:

```bash
packaging/homebrew/update-cask.sh 0.1.0 "$(cat dist/Portnado-v0.1.0-darwin-arm64.zip.sha256)"
```

Publish the cask into a local checkout of the tap without pushing:

```bash
HOMEBREW_TAP_DIR=/path/to/homebrew-tap HOMEBREW_TAP_PUSH=0 \
  packaging/homebrew/publish-tap.sh 0.1.0 "$(cat dist/Portnado-v0.1.0-darwin-arm64.zip.sha256)"
```

Tag releases can publish automatically when the repository secret
`HOMEBREW_TAP_TOKEN` is set to a token with contents write access to
`marcel-breuer/homebrew-tap`. The release workflow should never need
broader account permissions.

Portnado is not Developer ID signed or notarized. The Cask must not disable
Gatekeeper, remove quarantine, or claim Apple has verified the build.
