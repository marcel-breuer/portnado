# Troubleshooting

Start with:

```bash
portnado doctor
```

## Daemon Unavailable

Check whether the control socket exists and is user-scoped:

```bash
portnado status
portnado doctor
```

If launch at login was installed, inspect:

```text
~/Library/LaunchAgents/dev.portnado.daemon.plist
~/Library/Logs/Portnado/portnado.log
```

## Route Does Not Load

Confirm:

- the backend service is running,
- the route is approved and enabled,
- the request uses the `.localhost` host shown by `portnado list` or the menu bar,
- the high-port proxy is reachable on `127.0.0.1:4780`.

## Gatekeeper Warning

Portnado is currently unsigned and unnotarized. macOS may show an
unidentified-developer warning. Portnado does not disable Gatekeeper or remove
quarantine automatically.

## Uninstall

Preview cleanup:

```bash
portnado uninstall --dry-run
```

Repository `.portnado.yml` files are not removed.
