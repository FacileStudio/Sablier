# Changelog

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
