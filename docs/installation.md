# Installation

Portnado is distributed as one macOS archive:

```text
Portnado-vX.Y.Z-darwin-arm64.zip
```

The archive contains:

```text
Portnado.app
```

The app embeds:

```text
Portnado.app/Contents/Resources/bin/portnado
Portnado.app/Contents/Resources/bin/portnado-daemon
```

## Homebrew Cask

After a GitHub release is published and the tap is updated, install Portnado
with Homebrew:

```bash
brew install --cask marcel-breuer/portnado/portnado
```

The tap repository is `marcel-breuer/homebrew-portnado`, which Homebrew exposes
as the tap name `marcel-breuer/portnado`. The Cask installs `Portnado.app` into
`/Applications` and exposes the embedded CLI as `portnado`.

Portnado is not Developer ID signed or notarized. macOS may show an
unidentified-developer warning or say Apple cannot check the app for malicious
software on first launch. Portnado does not disable Gatekeeper or remove
quarantine automatically; approve the app manually only after verifying the
release source.

## Build Locally

```bash
PORTNADO_VERSION=0.1.0 make package-darwin-arm64
```

## First Setup

```bash
portnado doctor
portnado setup --dry-run
portnado scan --root "$PWD"
portnado list
```

Only apply launch at login after reviewing the setup plan.

```bash
portnado setup --launch-at-login --daemon-path /Applications/Portnado.app/Contents/Resources/bin/portnado-daemon --yes
```
