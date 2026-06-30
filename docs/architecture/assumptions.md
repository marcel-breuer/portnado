# Documented Assumptions

- The MVP target remains macOS 14 or newer on Apple Silicon.
- The project owner does not have a paid Apple Developer Program membership.
- Distribution is through a custom Homebrew tap, not official Homebrew Cask.
- Portnado may read local environment files only to resolve routing-relevant interpolation.
- Docker discovery uses the Docker CLI to remain compatible with Docker Desktop, OrbStack, and Colima where their CLI behavior matches Docker.
- Only published, host-reachable container ports are routable in the MVP.
- `.localhost` wildcard resolution usually works on macOS, but runtime diagnostics must verify it.
- Portless HTTP mode uses approved system setup; high-port mode remains the no-privilege fallback.
- Raw TCP routes require stable frontend ports.
- Discovery and route activation remain separate user-visible steps.
