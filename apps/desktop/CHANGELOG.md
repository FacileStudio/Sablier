# Changelog

Releases of the Sablier desktop app, tagged `desktop-v<version>`.

## [0.2.0] - 2026-11-03

### Added

- Self-update: Sablier checks GitHub 30 seconds after launch and then every 6 hours. When a new version exists it notifies once, and the tray menu offers "Install update X and restart". "Check for updates" runs a check on demand. Updates are signed; on Linux the deb or rpm installs through a `pkexec` prompt.

## [0.1.0] - 2026-11-03

### Added

- Tray timer, idle prompts, local window rules, start suggestions and heartbeats.
- "Open Sablier" opens the web app in its own window. On macOS the traffic lights sit in a draggable band above the sidebar.
