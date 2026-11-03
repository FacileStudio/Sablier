# Changelog

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
