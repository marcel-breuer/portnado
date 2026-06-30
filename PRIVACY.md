# Privacy

Portnado is designed to run locally.

For the MVP:

- no telemetry,
- no analytics,
- no cloud processing,
- no user accounts,
- no crash-report uploads,
- local database only,
- local logs only.

Portnado stores user state under `~/Library/Application Support/Portnado` and
logs under `~/Library/Logs/Portnado`. Homebrew Cask `zap` may remove those
paths, but repository `.portnado.yml` files are preserved.

Environment files may later be read locally to resolve routing-relevant
configuration, but raw environment values must not be logged or uploaded.
