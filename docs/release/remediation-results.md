# Release-Candidate Remediation Results

## Completed In This Pass

- Added local release artifact verification through
  `scripts/verify-release-artifact.sh`.
- Validated the Cask with `brew style --cask` and
  `brew audit --cask --strict` in a temporary local tap.
- Added PF anchor template rendering and tests without applying privileged
  system changes.
- Added focused tests for paths, deterministic IDs, Git root detection, marker
  files, LaunchAgent install/remove, doctor failure handling, and protocol
  response/params helpers.
- Rebuilt the local `Portnado-v0.1.0-darwin-arm64.zip` artifact and refreshed
  the Cask checksum.
- Ran a short packaged-daemon profile with temporary state.

## Still Not Verified

- Clean Apple Silicon Mac install.
- Installation from the external `marcel-breuer/tap` tap.
- Homebrew `uninstall` and `zap` against a real installed Cask.
- PF apply/rollback with administrator approval.
- Developer ID signing and notarization.
- Multi-hour daemon and menu bar resource profile.
