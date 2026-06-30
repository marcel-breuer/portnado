# Homebrew Tap Template

The intended tap is:

```text
marcel-breuer/homebrew-portnado
```

Initial bootstrapping does not require credentials. Copy `Casks/portnado.rb`
into the tap repository after building a release archive and calculating its
SHA-256.

Update the cask locally:

```bash
packaging/homebrew/update-cask.sh 0.1.0 "$(cat dist/Portnado-v0.1.0-darwin-arm64.zip.sha256)"
```

Later automation that pushes to the tap needs a repository token with contents
write access to `marcel-breuer/homebrew-portnado`. The release workflow should
never need broader account permissions.
