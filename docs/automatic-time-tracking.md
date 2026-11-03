# Automatic time tracking

Sablier records time with little or no manual input. AI agents record their own work as separate users. Tools send heartbeats for the human, and Sablier turns them into draft entries that confirm themselves if nobody reviews them. A desktop app adds a menu-bar timer, idle prompts and window rules.

## Goals

- Book every agent run in Sablier as an agent user with its own name and blob avatar, owned by the human it works for.
- Track human time without anyone pressing start. If the human forgets, the system still records the day.
- Show coworking time: the time a human and their agent worked at once.
- Keep window titles and file paths on the device.

## Components

| Piece | Where | Job |
|---|---|---|
| Agent users | `apps/api/modules/agents`, `internal/agentwork` | Agent accounts, agent sessions, paired report |
| Heartbeats and drafts | `apps/api/modules/heartbeats` | Heartbeat ingest, drafts computed on read, confirm, dismiss, auto-confirm |
| Desktop app | `apps/desktop` | Tray timer, idle prompts, window rules, suggestions, heartbeats |
| `sablier heartbeat`, `sablier hooks` | sablier-cli | Heartbeats from Claude Code hooks, git and scripts |
| Bulle runner | Bulle `apps/api/internal/runner/timetracking.go` | Books each Bulle run, sends a presence heartbeat per message |
| Mycelium sessions | `modules/antenne/agent_sessions.go` | Books Claude Code sessions from the Antenne bus to an agent user |

## Agent users

An agent user is a row in `users` with `kind = 'agent'`, `owner_user_id` set to its human and a unique `(owner_user_id, agent_key)`. It has a synthetic email, no password and no identity, so it can never sign in. Its rate is 0, so it never reaches an invoice. `agent_avatar_url` holds its blob, for example `https://blobatar.dev/avatar/Jarvis?gen=2`. Creating an agent adds it to each space its owner belongs to.

`POST /api/agents/sessions` records one span of agent work. The authenticated user is the owner. The call upserts by `facile_id`, so retries are safe. It replaces the whole entry, so callers send every field each time.

```json
{
  "agent_key": "bulle",
  "name": "Jarvis",
  "avatar_url": "https://blobatar.dev/avatar/Jarvis?gen=2",
  "facile_id": "bulle_run_<run id>",
  "project": "Bulle",
  "started_at": "2026-11-03T01:00:00Z",
  "stopped_at": "2026-11-03T01:30:00Z",
  "active_seconds": 1500,
  "cowork": true
}
```

Sablier picks the project in this order:

1. `project_id`
2. A project whose name matches `project`, ignoring case
3. The project of the owner's entry running at `started_at`
4. The shared "Agent work" project

The gap between wall time and `active_seconds` is stored as paused time. Bulle uses it to leave out time spent waiting on an approval.

`GET /api/agents` lists the caller's agents. `GET /api/agents/paired?from=&to=` returns, per agent:

- `agent_seconds`: the agent's active time
- `cowork_seconds`: the part of that time with a human in the session
- `paired_seconds`: the overlap between the owner's own entries and the agent's entries

The dashboard shows these in the Agents card for the current week.

Mycelium already sends `agent_session` events for Claude Code. They now go to an agent user named after the agent, such as "Claude". A one-time backfill moves older entries whose description starts with `mycelium: ` to that user. The enveloppe event contract is unchanged.

## Heartbeats and drafts

`POST /api/heartbeats` takes up to 500 heartbeats:

```json
{"heartbeats": [{"at": "2026-11-03T01:05:00Z", "source": "claude-code", "entity": "/home/me/Bulle", "project": "Bulle", "write": false}]}
```

Sources are `claude-code`, `editor`, `desktop`, `git`, `bulle` and `cli`. `entity` is a path or an app name, never a window title.

Sablier computes drafts when they are read, so it stores no draft rows that could go stale:

1. Merge heartbeats closer than 15 minutes into one span. A span ends 2 minutes after its last heartbeat.
2. Remove the time already covered by the user's entries (a running entry counts up to now) and by dismissed ranges.
3. Split a span where the resolved project changes. Drop pieces shorter than 1 minute.

`POST /api/heartbeats/drafts/confirm` creates a normal entry and refuses an overlap with 409. `POST /api/heartbeats/drafts/dismiss` hides a range. An hourly worker confirms drafts that ended more than 48 hours ago and have a project, for users with `auto_confirm_drafts` on, which is the default. The same worker deletes heartbeats older than 90 days. The dashboard lists today's and yesterday's drafts, and the profile page has the auto-confirm toggle.

## Desktop app

The tray app reads `~/.sablier.yml`, which `sablier login` writes. It does the following:

- Shows the running timer in the macOS menu bar, and in the tooltip and first menu item on Linux.
- Starts, stops and pauses the timer from the tray menu.
- Detects idle time after 3 minutes without input. Back from 5 minutes to 10 hours away with a timer running, it asks: Keep, Discard (stop the timer at the start of the idle time), or Discard and continue.
- Matches the active window against `~/.config/sablier/rules.yml`. After 5 minutes on one project with no timer, it suggests starting one. Rules with `auto: true` start it without asking.
- Opens the Sablier web app in its own window from "Open Sablier" and "Review suggestions". On macOS the window has no title bar: the traffic lights sit in a 40 px band above the sidebar, which drags the window. The window marks the page with `data-desktop`, and only that page can call `start_dragging`, nothing else. Linux keeps the native title bar.
- Sends `{at, source: "desktop", entity: <app name>, project}` heartbeats, at most one per app every 2 minutes, with an offline queue.

| Platform | Idle | Active window |
|---|---|---|
| macOS | `user-idle` | `active-win-pos-rs`; titles need the Screen Recording permission |
| Linux X11 | x11rb | x11rb |
| GNOME Wayland | Mutter `IdleMonitor` over D-Bus | the "Focused Window D-Bus" GNOME extension |
| KDE Wayland | `ext-idle-notify-v1` | kdotool |
| Hyprland | `ext-idle-notify-v1` | Hyprland socket |
| Other Wayland | `ext-idle-notify-v1` | none: the menu says window rules are unavailable |

Bundles are dmg for macOS and deb or rpm for Linux. There is no AppImage build, because an AppImage cannot read Wayland activity.

## Setup on a new machine

```sh
sablier login
sablier hooks install claude
sablier hooks print git >> .git/hooks/post-commit
```

Then install the desktop app and turn on start at login. Bulle reports runs on its own once `SUITE_SABLIER_URL` is set. Set `BULLE_SABLIER_TIME_TRACKING=false` to turn reporting off.

## Constants and prior art

| Constant | Value | Source |
|---|---|---|
| Merge gap | 15 min | WakaTime keystroke timeout, https://wakatime.com/faq |
| Client throttle | 2 min per entity unless a write | https://wakatime.com/help/creating-plugin |
| Idle detection | 3 min | aw-watcher-afk, https://github.com/ActivityWatch/aw-watcher-afk |
| Idle prompt window | 5 min to 10 h | Timing, https://timingapp.com/help/faq |
| Idle choices | keep, discard, discard and continue | Toggl Track, https://support.toggl.com/en-us/article/toggl-track-desktop-app-for-windows-5w1y5/ |
| Keep titles local | rules run on the device | Clockify auto tracker, https://clockify.me/help/apps/auto-tracker |
| Wayland backends | per-compositor tiers | awatcher, https://github.com/2e3s/awatcher; protocol support https://wayland.app/protocols/ext-idle-notify-v1 |
| Heartbeat shape | entity, project, is_write, time | https://wakatime.com/developers, ActivityWatch https://docs.activitywatch.net/en/latest/buckets-and-events.html |

## Open items

- The owner cannot edit or delete an agent's entries yet.
- The desktop app is released by a `desktop-v*` tag (`.github/workflows/desktop.yml`): a universal dmg plus deb and rpm. The dmg is not signed or notarized, and there is no updater yet. Both need Apple and Tauri signing secrets.
- The macOS build and the GNOME, KDE and Hyprland window backends have not been run on real desktops.
- Token counts sent with agent sessions are accepted but not stored.
