# Changelog

Releases of the Sablier native apps (desktop tray app and Android), tagged `desktop-v<version>`.

## [0.2.2] - 2026-11-03

### Fixed

- "Open Sablier" and "Review" no longer show a 404. The CLI stores `server_url` with an `/api` suffix, so the window opened `/api/...`; the app now strips it.

## [0.2.1] - 2026-11-03

### Fixed

- The release builds the Android APK again (`setup-android` v4). 0.2.0 was tagged but never published, so 0.2.1 is the first release with the Android app, self-update and the onboarding listed under 0.2.0.

## [0.2.0] - 2026-11-03

### Added

- Self-update: Sablier checks GitHub 30 seconds after launch and then every 6 hours. When a new version exists it notifies once, and the tray menu offers "Install update X and restart". "Check for updates" runs a check on demand. Updates are signed; on Linux the deb or rpm installs through a `pkexec` prompt.
- Android: `Sablier-android-universal.apk`, signed, with its SHA-256. It opens on an hourglass onboarding and signs in with Facile in the system browser.

## [0.1.0] - 2026-11-03

### Added

- Tray timer, idle prompts, local window rules, start suggestions and heartbeats.
- "Open Sablier" opens the web app in its own window. On macOS the traffic lights sit in a draggable band above the sidebar.
