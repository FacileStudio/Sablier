# Changelog

## [Unreleased]

### Changed

- `@facile/muse` v0.14.1 to v0.14.2. The user menu stays inside the viewport at the right edge of the header.
- The API tokens section drops the expiry field: Sablier tokens do not expire.
- The user menu gains a Subscription row that opens the plans on Facile (`https://sso.facile.studio/spaces`).

## [1.13.0] - 2026-11-07

### Changed

- `@facile/muse` v0.13.0 to v0.14.1. The user menu is one options object on the sidebar and the mobile bar: Facile account, Downloads (browser only), Settings, language, theme and Sign out. Settings and Sign out are no longer entries of the app's own list.
- Settings use muse's `SettingsLayout` in the order Profile, Preferences, Time tracking, Integrations, API tokens, Advanced. Profile is read-only and links to the Facile account; the identity colour stays. Rates and the auto-confirm switch moved to Time tracking. The Appearance tab became Preferences (language and theme), the API tab became API tokens with `TokensSection`, and the Advanced danger zone uses `DangerZone`. The in-settings Sign out is gone.
- `/settings/appearance` redirects to `/settings/preferences` and `/settings/api` to `/settings/tokens`.
- Choosing a theme no longer fails when the browser blocks local storage.

### Removed

- The local display name edit and the profile meta card, and the unused strings behind them.

## [1.12.1] - 2026-11-07

### Fixed
- `porte` v0.12.0 to v0.12.1: the GDPR erase no longer fails on a database without `porte_magic_links`, which is Sablier's production schema.

## [1.12.0] - 2026-11-07

### Added
- GDPR deletion: Sablier polls Registre's deletion list and accepts its signed `user.deleted` push at `POST /api/registre/deletions`. A deleted person loses their account, personal projects with their tasks and sessions, every time entry, their agent users, heartbeats, dismissed drafts, push subscription and memberships. A space project they owned passes to the space's longest-standing admin, else its longest-standing member; tasks they were assigned become unassigned. The receiver's cursor lives in the new `directory_deletion_state` table.
- `GET /version` serves the `APP_VERSION` build arg.
- `users.locale` stores the Registre locale, `users` answers it as `locale`, and the web-push reminder follows it.
- The client speaks French, English, Spanish and German, from the Registre profile locale, then the picked language, then the browser. Dates, numbers and currency follow it.
- The landing page uses muse's `LandingHeader` and `LandingFooter`.
- `.github/workflows/ci.yml` runs gofmt, vet, build, the Go tests on PostgreSQL, the client check, tests and build, and filet.

### Changed
- `porte` v0.8.0 to v0.12.0, `porte/directory` v0.1.0 to v0.3.0, `tronc` v0.16.0 to v0.17.1, `@facile/muse` v0.11.0 to v0.13.0.
- Sign-in goes through Registre only. The login page is muse's `AuthScreen`, every "Sign in with Facile" button is `SignInWithFacile`, and the API refuses to boot in production without `OIDC_ISSUER`.
- An email that already has an account (matched ignoring case) links to a Registre identity only when Registre verified it and the account holds no other subject of the same issuer.
- Security headers come from tronc: anti-framing is enforced, the full CSP is report-only. The client's inline scripts are hashed, and the connect list allows Ardoise's catalog, the icon API and the browser Journal origin.
- The Dockerfile build arg is `APP_VERSION` (was `VERSION`); compose passes it through.
- The task status dots and the sign-in screen carry no border.

### Removed
- `POST /api/auth/register`, `POST /api/auth/login`, the register and password forms, `SSO_ONLY` and the password half of `PATCH /api/users/me`. Clients that signed in with the account password must switch to an API token (`POST /api/users/me/api-token`).
- The unused `internal/authcrypto` package.

### Migrations
- `porte_login_codes.provider` and `.subject`, `users.locale` and the `directory_deletion_state` table, all idempotent and applied at boot.

## [1.11.0] - 2026-11-06

### Added
- The landing page has a "Pricing" section: muse's `PricingPlans` shows the live Facile plans from Ardoise's catalog. It renders nothing while the catalog is empty or unreachable.
- The landing page links to `/downloads` with muse's `DownloadCta` (macOS, Linux, Android).

### Changed
- `@facile/muse` v0.9.0 to v0.11.0.

## [1.10.0] - 2026-11-03

### Added
- Opus sync v2: projects follow their space both ways (Registre space id, or the owner's personal space), task description, due date and priority from Opus show read-only on the project page, and time entries on synced tasks go to Opus as created, updated and deleted events.
- Reconcile at boot and on `POST /api/antenne/sync`: same-named projects and tasks in the same space merge onto the smaller facile id.

### Fixed
- Projects created in Opus no longer all land in the personal view of user 1.
- Task moves between projects in Opus now apply in Sablier, and the local " (2)" name suffix is never sent back to Opus.
- Outbound events keep their order, and an older update can no longer overwrite a newer one.

## [1.9.0] - 2026-11-03

### Added
- The sidebar user card opens a menu: Settings, Download the apps (web only) and Sign out. The settings cog is gone. Uses `@facile/muse` v0.9.0.
- The client runs inside the desktop app: it signs in with Facile through the system browser and a `studio.facile.sablier://auth` link, like Android, and keeps the token in `~/.sablier.yml`.

### Changed
- CORS also allows `tauri://localhost` and `http://tauri.localhost` while the mobile sign-in is on.

## [1.8.1] - 2026-11-03

### Changed
- Agent avatars (the Bulle blobs) render without the colored disc behind them, and a bit bigger, everywhere an avatar shows: dashboard, projects, spaces, users and the agents card.

## [1.8.0] - 2026-11-03

### Added
- Android app (Capacitor, `apps/client/android`, appId `studio.facile.sablier`). It signs in through porte's mobile flow: PKCE in the system browser, a `studio.facile.sablier://auth` deep link, then a bearer token.
- On the native app, `/` never shows the landing page. Signed-out users get a four-step onboarding at `/welcome`, with an hourglass that drains one step at a time, ending on "Sign in with Facile". Signed-in users go to `/dashboard`.
- `/downloads` lists the macOS dmg, the Linux deb and rpm, and the universal Android APK with its checksum, and picks the visitor's OS.
- `OIDC_MOBILE_REDIRECT` (default `studio.facile.sablier://auth`). While OIDC is on, CORS also allows `https://localhost`, the Android app's origin.
- The client has unit tests (vitest): downloads, platform helpers, onboarding steps.

### Changed
- Bump porte to v0.8.0 for the mobile sign-in. The web cookie flow is unchanged.
- The viewport uses `viewport-fit=cover` so safe-area insets apply on phones.

## [1.7.1] - 2026-11-03

### Changed
- Spaces use the share-circle icon (`icons.space` from `@facile/muse` v0.8.1, which this release adopts) in the navigation, the space list, its empty state, the space header and the "Gérer dans Registre" buttons.

## [1.7.0] - 2026-11-03

### Added
- Desktop app (`apps/desktop`, released separately as `desktop-v0.1.0`): "Open Sablier" and "Review suggestions" open the web app in its own window instead of the browser. On macOS the window has no title bar, and the traffic lights sit in a draggable band above the sidebar.
- `.github/workflows/desktop.yml` builds a universal macOS dmg and Linux deb and rpm packages on `desktop-v*` tags and attaches them to a GitHub release.

### Changed
- When the desktop app marks the page with `data-desktop="macos"`, the sidebar column starts 40 px lower to leave room for the traffic lights. Browsers are unaffected.

## [1.6.0] - 2026-11-03

### Changed
- Cards, popovers and the login tab group lose their borders, matching `@facile/muse` v0.8.0, which this release adopts.

## [1.5.0] - 2026-11-03

### Added
- Shared spaces come from Registre. At boot Sablier imports its spaces into Registre once, retrying until Registre answers, then mirrors Registre's snapshot at boot, every 5 minutes and within seconds of a nudge. Members Registre adds are provisioned before their first sign-in, with their Registre name and picture. Agent users stay Sablier's own: they are never imported and never removed by a sync.
- `POST /api/registre/sync`, Registre's anonymous nudge. It answers `202` and is mounted only when `OIDC_ISSUER` is set.
- `spaces.registre_space_id`, nullable and unique, links a space to Registre. `Space` responses carry it.

### Changed
- Space pages are read-only and link to `https://sso.facile.studio/spaces` to create a space, invite, change a role, remove a member or leave.

### Removed
- `POST /api/spaces`, `PUT` and `DELETE /api/spaces/{id}`, `POST /api/spaces/{id}/leave`, `POST /api/spaces/{id}/members`, `PUT` and `DELETE /api/spaces/{id}/members/{memberId}`, and the new-space, space settings and member management pages. Registre has no space deletion yet, so neither does Sablier.

## [1.4.0] - 2026-11-03

### Added
- Agent users: `users.kind`, `owner_user_id`, `agent_key` and `agent_avatar_url`. An agent has its own name and blob avatar, belongs to one human, cannot sign in and has rate 0.
- `POST /api/agents/sessions` records an agent work span, upserted by `facile_id`. `GET /api/agents` lists the caller's agents. `GET /api/agents/paired` reports agent, cowork and paired seconds.
- `time_entries.cowork` marks agent time spent with a human in the session.
- Heartbeats: `POST /api/heartbeats`, plus drafts computed on read (`GET /api/heartbeats/drafts`, `confirm`, `dismiss`) and `GET`/`PUT /api/heartbeats/settings`.
- An hourly worker confirms drafts older than 48 h that have a project, and deletes heartbeats older than 90 days. Users can turn auto-confirm off on their profile.
- Dashboard: an Agents card for the week and a Suggested list of drafts.
- `apps/desktop`: a Tauri tray app with a menu-bar timer, idle prompts, local window rules, start suggestions and heartbeats.
- `docs/automatic-time-tracking.md`.

### Changed
- Mycelium `agent_session` events are booked to an agent user named after the agent instead of the human. Boot moves older `mycelium: ` entries to that user.
- Users, entries and space members carry `kind`, and the client shows an agent badge.
- `/api/auth` sessions, password login and OIDC refuse agent users, and the user count leaves them out.

## [1.3.0] - 2026-11-03

### Added
- Space member lists show each member's Registre profile picture, with initials when there is none.
- Boot backfills `oidc_picture_url` as `<issuer>/avatars/<subject>` for every identity under the configured issuer, so users who have not signed in since get their picture. The issuer comes from `OIDC_ISSUER`.

### Changed
- Bump muse to v0.7.3: the sidebar user card falls back to initials when the picture URL answers 404. `ColorPicker` is now `SwatchPicker`.
- The profile page links to `https://sso.facile.studio/account` to change the picture.

### Removed
- `POST` and `DELETE /api/users/me/avatar` and the upload UI. Pictures are set in Registre. Uploaded files and `avatar_upload_path` are kept and still render for accounts with no Registre identity.

## [1.2.0] - 2026-11-02

### Added
- `OIDC_MACHINE_AUDIENCE` and `OIDC_SCOPE_PREFIX`: Sablier accepts a delegated bearer JWT addressed to it, and gates it by method (`sablier.read` for reads, `sablier.write` for writes).

### Changed
- Bump porte to v0.7.0.
