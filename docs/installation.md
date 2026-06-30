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

The Cask template is:

```text
packaging/homebrew/Casks/portnado.rb
```

It installs `Portnado.app` into `/Applications` and exposes the embedded CLI as
`portnado`.

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
