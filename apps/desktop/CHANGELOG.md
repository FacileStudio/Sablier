# Changelog

Releases of the Sablier native apps (desktop tray app and Android), tagged `desktop-v<version>`.

## [Unreleased]

## [0.4.0] - 2026-11-13

### Added
- Windows: `Sablier-windows-x64-setup.exe`, an NSIS installer that self-updates like the other desktop builds. It is not code-signed, so SmartScreen warns on the first run.

### Security
- Android keeps its bearer token in the Keystore through `capacitor-secure-storage-plugin` instead of plain `@capacitor/preferences`. The first launch of the updated app moves an existing token across and deletes the Preferences key, so nobody is signed out; signing out clears both.
- The desktop app keeps its bearer token in the OS keyring (Keychain, Credential Manager, Secret Service) under `studio.facile.sablier` / `session`, the entry `sablier` CLI 0.6.0 uses, so the window, the tray and the CLI still share one sign-in. A token in `~/.sablier.yml` moves to the keyring on the first launch and leaves the file only once the keyring returned it. Without a usable keyring, such as Linux with no Secret Service, it stays in `~/.sablier.yml` (mode 0600). Signing out clears both.

### Changed
- The bundled app carries every Sablier web client change since 0.3.0, matching the web app at Sablier 1.22.0.

## [0.3.0] - 2026-11-03

### Changed
- The window shows the Sablier app bundled inside the desktop app instead of loading the website, which fixes the 404 on "Open Sablier".
- Launching the app opens its window: the onboarding and "Sign in with Facile" when signed out, the dashboard when signed in. Starting at login stays in the tray. Launching it again while it runs brings the window back.
- "Sign in" in the tray opens the in-app sign-in. Signing in there also signs in the tray and the `sablier` CLI, which share `~/.sablier.yml`.

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
