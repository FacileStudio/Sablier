# Plan: multi-space work and smarter attribution

Status: approved by the user on 2026-11-07. Ships as the next Sablier minor release, plus a Bulle minor release.

## Goals

1. Start a session in any space without switching spaces first.
2. Agent sessions and drafts arrive with the right space, project and task already filled in. Sablier books them on its own when it is sure, and asks the user when it is not.
3. Bulle does all the AI work. Sablier never holds an AI key.

## Decisions

| Topic | Decision |
| --- | --- |
| Space model | No global switch. The default view shows every space you can see. A space filter narrows the dashboard and reports. This is the pattern Agenda adopted in v0.9.0 (commits `6d323f6`, `5b968d5`). |
| Starting a session | One search field (also `Ctrl+K`). Results are tasks grouped by space, then project. Recent tasks and Opus tasks come first. Choosing a task also chooses its project and space. |
| Attribution | Layered. Explicit ids first, then hard facts (Opus task, Registre space), then an AI choice among that space's tasks. |
| Who resolves | Sablier runs the resolver. Bulle answers the AI question through its API (`POST /api/ai/decide`, wrapping `core/jev`). A Sablier deployed without the rest of the suite calls Jev itself, configured by its own env vars (A5). |
| Approval | Confident results book on their own and show a "why" chip. A confident result is an id match, or an AI confidence of at least the threshold (default 0.8, a user setting). Anything below the threshold waits in Review, prefilled with the best guess. |
| AI off | The AI step is skipped when no backend is configured or the call fails. The rules still run, and the result goes to Review. |
| Rollout | Every new field is optional, so an old Bulle keeps working against the new Sablier. Release Sablier first, then Bulle. |

Prior art:

- Toggl groups the project picker by workspace, so picking the project picks the workspace: https://support.toggl.com/en-us/article/how-are-time-entries-assigned-to-the-correct-workspace-18u1vve/
- Toggl runs rules before suggestions and can auto-assign the top suggested task: https://docs.toggl.com/my-activity
- Rize combines rules, task-tracker context and AI, and auto-tags only above a confidence threshold: http://docs.rize.io/automatic-tracking/time-entry-suggestions-settings
- In Timing, rules are not retroactive, and manual reassignment always wins over rules: https://timingapp.com/help/rules

## Track A: Sablier API

### A1. Visibility and the all-spaces view (fixes a security bug)

`projects.listProjects` and `timeentries.listEntries` / `listRunningEntries` filter by `?space_id=` without checking that the caller is a member of that space. Any signed-in user who knows a space's UUID can read its data.

- Move `visibleProject` out of `internal/agentwork/target.go` into one shared helper, for example `schemas.VisibleSpaces(tx, userID)`. Use it in all three list functions.
- Add `space_id=all` as the default for the client. It returns rows whose `space_id IS NULL`, plus rows in spaces the caller belongs to.
- A `space_id` the caller is not a member of returns 404.
- Use `CAST(space_id AS TEXT)`: `space_members.space_id` is text, while the other `space_id` columns are uuid. Smoke-test every new query on Postgres. SQLite does not catch this mismatch.

### A2. The server derives the space

`startTimer` and `createEntry` copy the `space_id` the client sends. Instead, set `entry.space_id = project.space_id` and ignore the client value.

### A3. A task search endpoint

Add `GET /api/tasks/search?q=&limit=`. It searches the visible tasks and returns `{task, project, space, last_used_at, opus}`, ranked as follows:

1. Your recent tasks, by the date of your latest entry.
2. Opus-linked tasks that are open.
3. Substring matches on the task name, then on the project name.

### A4. The attribution resolver

Add a new package `internal/attribution`. It replaces the matching steps in `agentwork.ResolveTarget` and also backs drafts.

Its input carries these facts, all optional:

- `project_id`, `task_id`
- `opus_task_ids[]`: Opus facile ids, which map to `tasks.facile_id`
- `opus_project_id`
- `registre_space_id`: maps to `spaces.registre_space_id`
- `opus_workspace`
- `title`, `project_hint`, `entity`

It returns `{project, task, space, source, confidence, needs_review}`. The source is one of `explicit`, `opus`, `name`, `running`, `ai` or `fallback`.

The resolver tries each step in order:

1. An explicit visible `task_id` or `project_id`. Confidence 1.
2. An Opus task id that maps to a visible task. Confidence 1. When several match, pick the one the run updated, otherwise the first.
3. The project hint matches a project name inside the resolved space, or across all visible projects. Confidence 0.9.
4. The owner's own entry that was running when the session started. Confidence 0.85, the current behaviour.
5. Ask Bulle (A5). The question is a Jev `choice` whose options are at most 30 candidate tasks, taken from the resolved space or from all visible projects when the space is unknown. Confidence is the value Jev returns.
6. Fallback: the existing "Agent work" catch-all, with `needs_review = true`.

`needs_review = confidence < user threshold`.

### A5. The decider: Bulle first, direct Jev for standalone installs

Add `internal/decide` with one interface, `Decide(ctx, userEmail, state, questions) (answers, error)`. Both backends speak the same Jev contract `{state, questions}` → `{answers}`, so they share one set of types. Each call has a 5 s timeout. Any error is treated as "no answer".

The backend is picked once at boot:

1. **Bulle**, when `BULLE_URL` is set. It calls `POST {BULLE_URL}/api/ai/decide` with `BULLE_SERVICE_KEY` (the auth choice is still open, see "Open"). This is the suite default, so Bulle stays the one place that does AI.
2. **Direct Jev**, when `BULLE_URL` is unset and `JEV_API_KEY` is set. This is for a Sablier deployed without the rest of the suite. It uses these env vars, with the same meaning as in Bulle's `core/jev`:
   - `JEV_API_KEY`
   - `JEV_ENDPOINT`: `systemone` (default) or `decisions`
   - `JEV_BASE_URL`: default `https://api.typesafe.ai`
   - `JEV_MODEL`: default `jev-latest`
3. **Off**, when neither is set.

Bulle wins when both are set. A failed Bulle call does not fall back to direct Jev. An install that chose Bulle has chosen where its data goes, and a silent second route would break that.

The direct client is a small copy of the request and response types from `core/jev`, about 80 lines in `internal/decide/jev.go`. Importing `github.com/FacileStudio/Bulle/core` would vendor the whole Bulle core for three JSON fields.

Log the chosen backend at boot. Document the five env vars in `apps/api/.env.example`, the root `.env.example` and `docs/configuration.md`.

The state contains only the session title, the project hint, the entity and the candidate labels in the form `Space › Project › Task`. It never contains message content.

### A6. Schema (migration on startup)

Add these columns to `time_entries`:

- `attribution_source text`
- `attribution_confidence real`
- `needs_review bool default false`

Reports still count entries flagged `needs_review`, so no time is lost.

Add this column to `users`: `review_threshold real default 0.8`.

### A7. Agents and heartbeats contracts

- `POST /api/agents/sessions` accepts the A4 facts.
- `POST /api/heartbeats` accepts `registre_space_id`, `opus_task_id` and `title`.
- Drafts call the resolver to prefill project and task, and expose `suggestion: {task_id, project_id, space_id, source, confidence}`.
- The hourly auto-confirm only confirms drafts whose suggestion has `needs_review = false`.

### A8. Review endpoints

- `GET /api/review`: entries flagged `needs_review` for the caller and the caller's agents, plus drafts.
- `POST /api/review/{entry_id}`: takes `{task_id}` to reassign the entry, or an empty body to accept it as it is. Either way, it clears the flag.

### Track A as built (2026-11-07)

The contract above holds. These are the additions and the points the plan left open:

- `space_id` semantics apply to `GET /api/projects`, `GET /api/time-entries` and `GET /api/time-entries?running=true`. `GET /api/time-entries/running` is the caller's own timer and takes no `space_id`. `running=true` now also honours `project_id` and `user_id`.
- `POST /api/projects` with a `space_id` you are not a member of returns 404. Start, create and update of an entry refuse a project you cannot see with 400.
- Entry rows (lists, running list, review, agent session response) carry `space_id`, `project_name` (lists and review), `attribution_source`, `attribution_confidence` and `needs_review`. `attribution_source` is null on entries a person booked.
- A seventh source, `manual`, marks an entry settled in Review (accepted or reassigned). A replayed agent session never moves a `manual` entry.
- `POST /api/agents/sessions` also accepts `opus_updated_task_id` (the Opus task the run updated, preferred over `opus_task_ids` order) and `entity`. Both are optional.
- Heartbeat `title` is optional, like `registre_space_id` and `opus_task_id`.
- `opus_workspace` maps to a member space by name, used only when `registre_space_id` is absent.
- A fallback always needs review, whatever the threshold. A draft whose resolver found nothing has `suggestion: null`. A draft with a suggestion but no `project_id` of its own gets the suggestion's project as `project_id`.
- `review_threshold` lives in `GET/PUT /api/heartbeats/settings` next to `auto_confirm_drafts`. A PUT without it keeps the stored value; outside 0.5 to 1 is a 400.
- `POST /api/review/{entry_id}` accepts `{"task_id": n}`, `{}` or an empty body.
- `review_threshold` and `attribution_confidence` are `double precision`, not `real`: a `real` 0.8 reads back as 0.800000011920929.
- Bulle auth: `Authorization: Bearer <BULLE_SERVICE_KEY>` plus `X-Bulle-Acting-For: <email>`. Any non-200 (404 from an older Bulle, 402 `payment_required` from C4) is no answer. Drafts send one decide call per listing, one question per unplaced draft.

## Track B: Sablier client

- B1. Remove the sidebar space switcher (`selectSpace` with `window.location.reload()`). Remove the old `space-context` active-space state, and keep only a per-page filter, stored in the URL as `?space=`.
- B2. Add a space filter dropdown to the dashboard header: All, Personal, then each space. Use the same filter on the projects list.
- B3. Rebuild TimerControl as a command palette that opens with `Ctrl+K` or the play button. It uses one search field with results grouped by space, as in the approved mockup, and offers "+ New task" in a chosen project. The running timer shows its space with the user's space colour.
- B4. Replace the Suggested table with a Review inbox for drafts and flagged agent entries. Each row is prefilled with the guess and shows a chip such as "Opus task" or "AI 72 %". It has accept, change-task (the B3 picker), and dismiss actions.
- B5. Booked entries show a small source chip, so you can see why they landed where they did.
- B6. Add a threshold slider in Settings, next to the auto-confirm toggle.

Use muse components. Follow the Svelte 5 runes rules. The client has four languages (`src/lib/i18n`, since v1.12.0), so add every new string in all four. Base commit: `71e23f6` (v1.12.1).

## Track C: Bulle (separate repo, its own worktree off `origin/main`)

- C1. `POST /api/ai/decide`: it takes the Jev `{state, questions}` and returns the answers. It reuses `core/jev`, with the key and model in Bulle's config. It is rate-limited per caller and logs usage.
- C2. In `runner/timetracking.go` and `suite/sablier.go`, `AgentSession` gains these fields: `registre_space_id` and `opus_workspace` (from the session's space), `opus_task_ids` (collected from `opus_task` / `opus_update_task` / `opus_create_task` tool calls during the run), and `title`. The `recordPresence` heartbeat gains `registre_space_id`.
- C3. Tests: the Opus task ids collected from a fake run, and the decide handler with a stub Jev server.
- C4. Before merging, added 2026-11-07:
  - Rebase `feat/ai-decide` on Bulle v3.28.0. That release adds migration 00024 (`forfait_state`) and the plans-gate wiring in `main.go`, so expect a small conflict in `main.go`.
  - In `/api/ai/decide`, call `Gate.CheckAI(ctx, userID)` for the acting-for user, right after the service-key auth. Refuse with 402, code `payment_required`. Add a handler test for this case.
  - Sablier treats the 402 like any other non-200 and sends the session to Review.

## Auth for Sablier calling Bulle (settled 2026-11-07 with the Bulle session)

Sablier uses a service key plus an `X-Bulle-Acting-For: <email>` header, the Prisme pattern. Registre is not touched.

Bulle enforces the following:

- One key per calling service. Bulle stores an allow-list of callers, for example `BULLE_AI_CALLERS=sablier:<sha256 of key>`, and compares keys in constant time.
- The acting-for email must belong to an existing Bulle user. Otherwise Bulle answers 403.
- Bulle reads the acting-for header only on service-key requests. It is ignored on cookie or user-bearer requests and never widens them.
- Each caller has its own rate limit.
- Every decide call writes one audit log line with the caller, user id, question count, model and usage.
- Jev config reaches the API process the way `internal/disclosure` already does: `jev.Config{}` falls back to `TYPESAFE_API_KEY`. Expose `JEV_ENDPOINT`, `JEV_BASE_URL` and `JEV_MODEL` the same way, and document them.
- Runner edits stay in `runner/timetracking.go` plus the `RecordAgentSession` call site. The Bulle session is about to refactor the rest of `internal/runner`.
- The Bulle session owns Bulle releases. Merge to Bulle main and tell bulle-66, and they ship it in v3.28.0. Never tag Bulle from here.

Until that Bulle release, Sablier gets 404 from `/api/ai/decide`, treats it as "no answer", and routes the result to Review.

## Checks

- API: `go test ./...`, a Postgres smoke run for the new queries, and `filet check` if `filet.yml` exists.
- Client: `bun run check` and `bun test`.
- Bulle: its own gate (`mycelium flow run suite-check`).
- The smallest runnable check for the resolver is a table test with one case per step.
