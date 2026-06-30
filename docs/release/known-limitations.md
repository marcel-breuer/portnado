# Known Limitations

- A clean Apple Silicon Mac installation was not tested in this workspace.
- Homebrew Cask installation was templated but not installed from the external
  tap.
- `Portnado.app` was built and archived locally, but the app was not launched as
  an installed `/Applications` bundle in a clean user account.
- The release archive is not Developer ID signed or notarized.
- PF portless mode is previewed and documented, but managed PF apply/rollback is
  not implemented.
- LaunchAgent plist writing is implemented, but login-session launch behavior was
  not verified on a clean machine.
- Swift UI tests cover model/IPC encoding only; no automated menu bar UI
  interaction test exists yet.
- Short daemon profiling completed, but sustained multi-hour daemon and menu bar
  resource profiling remains a release-candidate follow-up.
- The aggregate Go coverage target is not yet at 80%; the current pass measured
  62.8%.
- Very long HOME paths can exceed macOS Unix socket path limits for the control
  socket; normal user home paths are shorter, but this needs clearer diagnostics.
