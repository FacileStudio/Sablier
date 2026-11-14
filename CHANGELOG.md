# Changelog

## [1.24.2] - 2026-11-14

### Security

- Browser session mark fix: porte v0.20.1 marks browser session tokens with `b~` instead of `b-`. A plain token is base64url, so 1 in 4096 CLI or API tokens started with `b-` and could pass the MCP confirm page as a browser cookie. People sign in once more in their browser before they can confirm an MCP action.
- porte v0.20.1, porte/mcpauth v0.3.1, tronc v0.23.1.

## [1.24.1] - 2026-11-14

### Changed

- The landing download section shows only the visitor's device (muse `DownloadSection` `mode="device"`), with an "Other platforms" link to `/downloads`. The downloads page still lists every platform and the APK steps.
- The hero download button points at `#download` on the same page.
- muse v0.23.0 and mise-en-scene v0.2.2. mise-en-scene needs no Vite `optimizeDeps` entry any more.

## [1.24.0] - 2026-11-14

### Changed

- Attribution AI goes through Bulle only, with `tronc/bulle` `DecideAI`. Sablier checks the owner's `ai.access` entitlement once per batch before any call; without it, the work waits in Review. Without both `BULLE_URL` and `BULLE_SERVICE_KEY`, the AI step is off.
- Only a browser session cookie can confirm or dismiss a pending MCP action (porte/mcpauth `RequireOwnSession`). A CLI session, bearer token, API key or delegated token gets 403. Cookies issued before this release are refused until the person signs in again.
- The pending page shows a translated warning on that refusal: sign in again in your browser; the mobile and desktop apps cannot confirm.
- tronc v0.23.0, porte v0.20.0, porte/mcpauth v0.3.0, porte/forfait v0.1.3. The MCP confirm page, state key, account check, pruning, waiting text and list pager use the shared kit helpers.
- `sablier_list_entries` and `sablier_list_projects` use the kit cursor. Old cursors are refused; clients start the list again.

### Removed

- The direct Jev backend and the `JEV_*` variables. Sablier no longer calls typesafe.ai or openrouter.ai.

### Fixed

- Account erasure deletes the person's and their agents' pending MCP actions with `ErasePending`, in the same transaction.
- A negative list offset is refused instead of panicking.

## [1.23.0] - 2026-11-14

### Added

- The landing page has an animated demo built with mise-en-scene v0.2.0: a timer starts on a project, counts up, stops, and the session joins the day's list.
- The landing page has a download section (muse `DownloadSection`) with the desktop, Android and CLI installs; the hero's download button scrolls to it.
- The Windows installer (`Sablier-windows-x64-setup.exe`, shipped since desktop-v0.4.0) is listed on the landing and downloads pages.
- The CLI install commands (install script, `facile install sablier`, `brew install FacileStudio/tap/sablier`) are listed on both pages.

### Changed

- muse v0.22.0: titles and body text use Magnetik.
- Landing copy is shorter, with four features: one timer everywhere, drafts from activity, agents, shared spaces.
- The downloads page uses muse `DownloadSection` instead of its own cards, and keeps the APK install steps.

## [1.22.4] - 2026-11-13

### Changed

- muse v0.21.0: icon buttons show their label as a muse tooltip on hover and keyboard focus.

### Fixed

- The task status dot, the icon-only rename and delete buttons for tasks and sessions, and the mobile timer button show a muse tooltip; the status dot's native `title` is gone.

## [1.22.3] - 2026-11-13

### Added

- The web client loads Vision's tracker (`https://vision.facile.studio/s.js`, with `data-mask`) on `sablier.facile.studio` only. `script-src` allows the Vision origin next to the inline script hashes.

## [1.22.2] - 2026-11-13

### Fixed

- Timer reminders: a push service that answered with an HTTP error was logged at ERROR every minute as `"error":null`, because only the transport error was logged and a failed send left `last_notification_at` untouched. The log now carries the push service's status, the first 256 bytes of its answer, the subscription id and the endpoint host. A 404 or 410 (the subscription is gone) deletes that `push_subscriptions` row, so it is never tried again. Any other failure keeps the row, logs at WARN and waits an hour before the next attempt, the same cadence as a delivered reminder.

## [1.22.1] - 2026-11-13

### Changed

- Compose healthchecks run every 30 s instead of every 5 s (db) and 10 s (api), which cuts the steady `docker exec` load on the host. `start_interval: 1s` during `start_period` keeps a fresh container's first passing check, and so Traefik's cutover on deploy, at about one second.

## [1.22.0] - 2026-11-13

### Security

- Android: the app keeps its bearer token in the Keystore through `capacitor-secure-storage-plugin` 0.13.0 instead of plain `@capacitor/preferences`. The first launch of an updated app moves an existing token into the secure storage and deletes the Preferences key, so nobody is signed out; when the Keystore refuses it, the token still serves that run and the Preferences copy stays for the next launch. Signing out clears both stores. The web build is unchanged, and the desktop app keeps its token in `~/.sablier.yml` (mode 0600), shared with the tray and the `sablier` CLI. The Android app ships with the next `desktop-v*` tag.

## [1.21.0] - 2026-11-13

### Changed

- Database: the schema is ordered goose SQL under `apps/api/migrations/`, applied at boot by `tronc/migrate` v0.1.1 behind a Postgres advisory lock. GORM `AutoMigrate` and the DDL that ran at every boot (the porte tables, `directory_deletion_state`, `mcp_actions`, the antenne outbox) are gone from the boot path; the GORM structs stay as the query layer. `scripts/check.sh` fails on an `.AutoMigrate(` outside a test file, and `schemas/columns_test.go` fails when a struct names a column the migrations do not create or a type its field cannot hold.
- Database: `00001_baseline.sql` is production's schema as `pg_dump --schema-only` printed it on 2026-11-13, leftover columns included; a database built from it dumps byte-identical to production. A database built before this release (Sablier's tables, no applied migration) is stamped at version 1 on its first boot with tronc/migrate's `baseline` command, so the baseline is recorded without running. `apps/api/schema.sql` is the normalized dump of the whole set. `migrations_test.go` pins the `mcp_actions` and `antenne_outbox` tables to the DDL of tronc/mcp and antenne-client.
- Boot: the data steps that still change rows run at every boot after the migrations, in the same order (`schemas.Reconcile`): the Registre picture backfill, the avatar upload path, missing user colours, legacy entry tasks, missing facile ids and Mycelium entries moved to agent users. The steps that only served databases older than production are removed, since production holds none of their input: the `nook_pool_*` column and `pool_*` table renames, the in-place conversion of the pre-v1.19.0 outbox, the copy of the pre-porte `sessions` and `api_tokens` tables, the `oidc_subject` move into `porte_identities` and the re-key of password identities.
- Tests: every database test runs on PostgreSQL through `internal/testdb.Migrated`, on a schema the migrations build. The in-memory SQLite helpers are gone, and with them `gorm.io/driver/sqlite` and `mattn/go-sqlite3` from `go.mod` and `vendor/`; `scripts/check.sh` fails on a non-PostgreSQL driver in `go.mod`.
- build: `github.com/FacileStudio/tronc/migrate` v0.1.1 with `pressly/goose/v3` v3.27.3; goose raises `segmentio/asm` from v1.1.3 to v1.2.1. Re-vendored.

### Added

- `docker run <image> migrate status` (also `up`, `version`, `baseline <n>`) answers without starting a server.
- Database: `00002_column_defaults.sql` gives `created_at`, `updated_at`, `joined_at` and `processed_at` a `now()` default on every table that has them, and the `users` columns porte/kit does not write the values `fillNewAccount` writes (`''` and `0001-01-01 00:00:00+00`), so a row inserted outside GORM no longer gets NULL there. It has a Down. Rolling back to v1.20.0 or older drops these defaults again, because AutoMigrate removes a default its struct tags do not declare, while the ledger still records version 2: before deploying this release again after such a rollback, run `DELETE FROM goose_db_version WHERE version_id = 2`. Rolling back past v1.19.0 is not supported: its outbox shape is no longer converted.

## [1.20.0] - 2026-11-12

### Added

- MCP tools `sablier_record_agent_session` and `sablier_record_heartbeat`, so an agent books its own working time over MCP. They run the same validation and storage as `POST /api/agents/sessions` and `POST /api/heartbeats`: a session is upserted by `facile_id`, and an exact repeat of a heartbeat changes nothing. Both are normal writes, with no confirmation. The REST routes stay for the CLI, the desktop app and other clients.

### Fixed

- A person who never opened Sablier got 401 on their first MCP call, so Bulle showed Sablier as "no account" until a web sign-in. Production set `OIDC_CLAIMS_SCOPE=roles`, and porte then refuses to provision from a token without a `roles` claim, which Registre leaves out of MCP resource tokens on purpose. Sablier never reads roles, so `OIDC_CLAIMS_SCOPE` is gone and the first MCP call creates the account again.
- An account re-created for a person Registre erased (a sign-in or MCP call racing the erasure) stayed until a restart. porte v0.19.0 and porte/directory v0.5.0: the deletion receiver erases such an account again on the next notice or poll (`Deletions.Holds`), and sign-in, MCP provisioning and the profile sync refuse an erased subject (`oidc.Deps.Erased`).

## [1.19.5] - 2026-11-11

### Fixed

- The favicon, touch and manifest icons, link preview image and web manifest come from the muse v0.19.0 brand generator, and the head carries the suite's canonical title, description, theme colours and Open Graph and Twitter tags. The landing page title and description match the suite catalog.

## [1.19.4] - 2026-11-11

### Security

- Client dependencies past their high advisories: `@sveltejs/kit` 2.70.4 and `vite` 7.3.8, with overrides pinning `devalue` 5.9.4, `nanoid` 3.3.20, `picomatch` 4.0.7, `postcss` 8.5.32 and `source-map-js` 1.2.2.
- The unused `apps/client/package-lock.json` is deleted. `bun.lock` is the only lockfile, and the stale npm lock kept old vulnerable versions in the scans.

## [1.19.3] - 2026-11-11

### Fixed

- AI attribution answered nothing through Bulle: every choice question went out without `criteria`, and the
  decision model behind `POST /api/ai/decide` refused it with 422 (Bulle answered 503). Each question now
  carries a criterion per option, plus a "None of these tasks" option, so an item with a single candidate task
  still asks a valid two-option choice. Direct Jev installs send the same shape.
- A batch of open items is split into calls of at most eight questions and under Bulle's 32 KiB state and
  128 KiB body limits. The hourly draft pass sent everything in one call and got 413.

## [1.19.2] - 2026-11-11

### Changed

- The MCP server and its confirm page run on tronc/mcp v0.22.0 (`NewApp`, `PendingPage`, `Digest`,
  `TargetIDs`) instead of local copies. MCP tool names, schemas and instructions are unchanged, and open
  pending deletions still confirm. The OpenAPI list schema is now named `PendingList_PendingItem`, and the
  confirm and dismiss result documents an optional `url`.
- Go 1.26.11 (the `go` directive and the builder image) clears the standard library advisories douane
  reports against 1.26.10.

## [1.19.1] - 2026-11-11

### Security

- `GET` and `PUT /api/settings` and every `/api/antenne*` route answer 403 unless the account's email is on the
  new `SABLIER_ADMINS` list. Empty, the default, closes them to everyone. Before, any account that could sign in
  through Registre could read the webhook secret and the Antenne bus secret, and repoint the webhook or the bus.
- Secrets are write-only. `GET /api/settings` returns `webhook_secret_set` instead of `webhook_secret_value`,
  and `GET /api/antenne` returns `antenne_secret_set` instead of `antenne_secret`. A `PUT` without the secret
  field keeps the stored one; an empty `webhook_secret_value` clears it.
- `GET /api/users` and `GET /api/users/{id}` show only the caller, its own agents and the people it shares a
  space with, without `locale` and `avatar_source`. Any other id answers 404. Before, every account saw every
  account's email and rate.
- `PATCH /api/users/me` no longer changes the email: Registre owns it, and the admin list compares it.
- `POST /api/notifications/test-broadcast` is removed. It took no authentication and pushed any text to every
  subscribed browser.
- `/api/users/me` carries `is_admin`, and the Integrations settings tab shows only to admins.

### Changed

- Licensed under the GNU Affero General Public License v3.0 only (`AGPL-3.0-only`). See `LICENSE`. The desktop shell, previously declared `MIT` in `Cargo.toml`, moves to the same licence; releases up to 1.19.0 keep the terms they shipped with.

## [1.19.0] - 2026-11-10

### Added

- `GET /api/me/invitations` (porte/kit `MountInvitations`, mounted while OIDC is on): the signed-in person's
  pending Registre space invitations, shown in the user menu through muse v0.18.0's `invitations` row.

### Changed

- Sign-in runs on `porte/kit`: `kit.NewStore` over the `users` table, `kit.Mount` and `kit.RequireAuth`.
  The stored email is now rewritten only when Registre marks it verified, the suite rule. Sablier keeps
  app-side the refusal of any sign-in that would reach an agent user, the decimal-string user id its handlers
  read, and the colour a new account gets.
- Erasure runs through the kit's `EraseAccount`, in one transaction under the sign-in lock. It now also deletes
  the person's pending MCP actions and every queued or recently sent bus event naming their email or subject.
- The bus moves to `antenne-client/go` v1.1.0: `envelope` replaces `enveloppe/go`, channel names come from
  `catalog`, and events queue in the library's outbox. Boot converts the existing `antenne_outbox` table in
  place and keeps its unsent rows. Every published event keeps its exact bytes (`modules/antenne/testdata/wire`).
  The client keeps dialling a bus that is down at boot instead of giving up. `ANTENNE_INSTANCE` is read as a
  synonym of `ANTENNE_URL`. Sent events stay in the outbox for 7 days, where they used to be deleted on send.
- The browser calls the API through `@facile/client` v0.1.1 and sends no bearer token; a `sablier.token` key
  left in `localStorage` is deleted on load. The Tauri and Android apps keep their bearer token in native
  storage, unchanged.
- `tronc` v0.20.0 to v0.21.0, `porte` v0.16.0 to v0.18.2 (`directory` v0.4.0), `@facile/muse` v0.17.1 to
  v0.18.0, Go 1.26.10, and the newest `x/crypto`, `x/text`, `go-jose` and `chi`.
- Gates: `filet.yml` is `preset: facile`, the pre-push hook runs `filet gate`, CI pins filet v0.19.1 and runs
  douane v0.7.1 with `-fail critical -scope prod`.

## [1.18.0] - 2026-11-10

### Added

- A remote MCP server at `POST /mcp` (stateless, spec 2026-07-28, on `tronc/mcp`), so any MCP client can track
  time in Sablier with a Registre token bound to `https://sablier.facile.studio/mcp` and the scopes
  `sablier.read` and `sablier.write`. Nine tools: `sablier_timer_status`, `sablier_start_timer`,
  `sablier_stop_timer`, `sablier_list_entries` (a range with totals per project), `sablier_add_entry`,
  `sablier_update_entry` (only the fields given change), `sablier_delete_entry`, `sablier_list_projects` (with
  tasks and spaces) and `sablier_weekly_summary` (per day, project, task and person, in the caller's timezone).
  Every tool calls the same service method as its REST route, so spaces and permissions match the API.
- Deleting an entry through MCP always needs the person: a client that can ask confirms in place, any other
  gets a link to the new `/pending` page, where only the person's own session can confirm. Pending actions live
  in the new `mcp_actions` table.
- A person's first MCP call creates their Sablier account, as a first web sign-in would (`porte/mcpauth`
  provisioning).
- `MCP_RESOURCE` (compose default `https://sablier.facile.studio/mcp`) turns the server on while OIDC is on.

### Changed

- `porte` v0.15.1 to v0.16.0, `tronc` v0.18.0 to v0.20.0, plus `porte/mcpauth` v0.2.0 and
  `modelcontextprotocol/go-sdk` v1.8.0.

## [1.17.0] - 2026-11-10

### Added

- An Ideas page, first in the sidebar and the mobile bar with the lightbulb icon and a badge for new ideas.
  Bulle suggests what to do next for Sablier. You generate ideas, run one in place, answer Bulle's approval
  cards, then follow the output links or continue in Bulle. A space picker scopes the ideas to a shared space.
- `/api/ideas*` proxies Bulle's Ideas API (`tronc/bulle`) with a Registre token exchange (`porte/exchange`,
  audience `bulle`). It needs `BULLE_URL`; Sablier already seals refresh tokens with `PORTE_TOKEN_KEY`.
- Ideas are an AI feature: a plan without `ai.access` sees an upsell page, and the API answers 402 without
  calling Bulle. Sablier now reads entitlements from the Registre snapshot (`porte/forfait`) and keeps them in
  the new `forfait_state` table.

### Changed

- `porte` v0.13.0 to v0.15.1, `tronc` v0.17.1 to v0.18.0 (additive: the `bulle` package), `@facile/muse`
  v0.16.1 to v0.17.1.

## [1.16.2] - 2026-11-08

### Fixed

- `docker-compose.yml` forwards `BULLE_URL`, `BULLE_SERVICE_KEY` and the `JEV_*` variables to the API container.
  Without them the AI step of attribution stayed off in production even with the keys set in Dokploy.

## [1.16.1] - 2026-11-07

### Changed

- `@facile/muse` v0.16.0 to v0.16.1: checkboxes, radios and switches keep their 44px touch target. Buttons use
  the current variant names (`outline` is `secondary`, `danger` is `ghost-danger`, icon buttons say
  `secondary`); nothing else changes on screen.
- `GET /version` serves the newest released `CHANGELOG.md` heading, baked into the binary at build time by
  `scripts/version.sh`. The `APP_VERSION` build arg is gone from the Dockerfile, compose and the image
  workflow, so a hand-set Dokploy env can no longer make it lag. The response shape stays `{"version":"1.16.0"}`.

## [1.16.0] - 2026-11-07

### Changed

- `@facile/muse` v0.14.4 to v0.16.0. The project icon picker is now muse's `IconPicker`, with the same
  1,451 linear Solar icons, labelled by its `Field`, with 44px cells; typing in the grid jumps to the search
  and Enter picks a single match. Icon names come from muse's `solarIconId`; Sablier's own picker, icon list
  and picker strings are gone.

## [1.15.0] - 2026-11-07

### Added

- The project icon picker offers every Solar icon in its linear style: 1,451 instead of 1,231. Names match
  Opus, so a project icon reads the same in both apps; names Solar has since renamed keep rendering.

### Fixed

- The project icon picker opens in the top layer above the project drawer, so the drawer no longer clips it
  or pushes it over the form. It stays inside the viewport, opens upward only when there is more room
  there, fits its grid to its width and closes on any click outside.

## [1.14.1] - 2026-11-07

### Changed

- `@facile/muse` v0.14.2 to v0.14.4: the language chosen in the user menu re-renders the whole UI without a reload, and a single 100% slice in the dashboard donut is drawn as a clean ring.
- The user menu's Subscription row drops its empty `plan` placeholder and shows only its label.

## [1.14.0] - 2026-11-07

### Security

- **Breaking:** personal projects, tasks and entries (no space) are private to their owner. Before, every
  signed-in account could read them. A project is yours when you own it, an entry when you or one of your
  agents booked it. Team work belongs in a Registre space.
- `?space_id=` on the project and time-entry lists checks membership: a space you are not in is a 404.
- Every `/api/projects/{id}` route, tasks included, answers 404 for a project you cannot see.
- Creating a project in a foreign space is a 404. Starting, creating or updating an entry on a project you
  cannot see is a 400. The server takes an entry's space from its project and ignores the client's value.

### Added

- Work across spaces without switching. The dashboard and projects page have a space filter (All, Personal,
  each space), kept in the URL as `?space=`. `space_id=all` lists personal rows plus every member space.
- A task palette (Ctrl+K / Cmd+K, or the play button) starts a session in any space. It searches tasks grouped
  by space and project, recent and Opus tasks first, and can create a task. `GET /api/tasks/search` backs it.
- Agent sessions and drafts are attributed in layers: explicit ids, Opus task ids, Registre space, project
  name, the running timer, then an AI choice among that space's tasks. Each entry records
  `attribution_source`, `attribution_confidence` and `needs_review`.
- The AI choice goes through Bulle (`BULLE_URL`, `BULLE_SERVICE_KEY`, `POST /api/ai/decide`), or straight to
  Jev for a Sablier without the suite (`JEV_API_KEY`, `JEV_ENDPOINT`, `JEV_BASE_URL`, `JEV_MODEL`). Bulle wins
  when both are set. With neither, the rules still run. Answers are cached for ten minutes, failures for one.
- A Review inbox (`GET /api/review`, `POST /api/review/{entry_id}`) replaces the Suggested table. Sure
  results book themselves; the rest wait there, prefilled with the best guess. Accepted entries are never
  moved again by a replayed agent session.
- A review threshold (0.5 to 1, default 0.8) under Time tracking settings, saved as `review_threshold`.
- `POST /api/agents/sessions` accepts `registre_space_id`, `opus_workspace`, `opus_task_ids`,
  `opus_updated_task_id`, `title` and `entity`. `POST /api/heartbeats` accepts `registre_space_id`,
  `opus_task_id` and `title`. Drafts carry a `suggestion`.

### Changed

- The sidebar space switcher and its page reload are gone.
- Each owner gets their own "Agent work" catch-all project instead of one shared project.

## [1.13.2] - 2026-11-07

### Security

- Registre tokens are sealed at rest. porte to v0.13.0: `porte_identities.access_token` and `refresh_token`
  are encrypted with AES-256-GCM under `PORTE_TOKEN_KEY`. Existing rows are sealed once, in the background,
  after boot, and the count is logged. An invalid `PORTE_TOKEN_KEY` stops the boot. With the key unset nothing changes.
- Losing `PORTE_TOKEN_KEY` makes the stored tokens unreadable: people sign in again, nothing else is lost.
  Do not roll back to a build older than porte v0.13.0 after the first reseal.

## [1.13.1] - 2026-11-07

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
