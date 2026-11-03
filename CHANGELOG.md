# Changelog

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
