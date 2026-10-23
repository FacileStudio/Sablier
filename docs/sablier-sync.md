# Opus sync

Sablier and Opus keep projects and tasks in step over the Antenne pool. Both
apps emit `project.{created,updated,deleted}` and `task.{created,updated,deleted}`
for local user changes and apply the same events from the other app.

## Rules

- **Identity.** Every project and task has a unique `facile_id` (`fac_` plus hex), minted at
  creation. Startup backfills rows that lack one. Inbound events match rows by `facile_id` only.
- **Stable ids only.** Rows are created only from ids Opus or Sablier mint today: 26 characters
  of `A-Z2-7`, `fac_` plus 20 hex digits, or 32 hex digits (Opus migration backfill). Numeric, slug and TS-era cuid2 ids never create a
  row, so replayed old events cannot add junk. Rows that already exist still match by exact id.
- **No loops.** Only events whose `app` is `Opus` are applied; `Sablier` and any other app are ignored. Applying an inbound event writes
  straight to the database and never queues an outbound event.
- **Idempotency.** Each applied event's `idempotency_key` is claimed in the ledger
  (`AntenneProcessedEvent`, pruned after 35 days) inside the same transaction as the change. A
  replay does nothing. A dropped event rolls the claim back, so it can be replayed. An event whose envelope
  `timestamp` is missing or older than the ledger retention is dropped, so a replay after the
  ledger is pruned cannot overwrite newer state.
- **Tasks need their project.** A task event whose `project_facile_id` matches no local project
  is dropped with a warning. No orphan task is created.
- **Updates for unknown rows.** `project.updated` for an unknown project is dropped; only
  `project.created` creates a project. `task.updated` for an unknown task creates it when its id
  is stable and its project is known.
- **Task names.** A task whose name collides (case-insensitive) inside its project gets a
  ` (2)`, ` (3)` suffix instead of being lost.
- **Deletes** are a no-op when the row is gone. Deleting a project removes its tasks. Time
  entries of a deleted task are kept and unlinked.
- **Fail soft.** With the pool down or unconfigured, local writes succeed and events wait in the outbox.
- **Extra fields.** Opus payloads carry `id`, `project_id`, `workspace_id`, `title`, `priority`,
  `due_date`, `actor_id`, `slug`. Sablier ignores them. The name is `name`, or `title` when
  `name` is empty.

Projects created from Opus belong to user 1 and have no space.

## Status mapping

Sablier statuses are `to-do`, `in-progress`, `in-review`, `done`. Opus statuses are workflow
column slugs, and Opus's default columns use the same four slugs, so Sablier sends its status
unchanged. Inbound, the value is lowercased and spaces or underscores become `-`, then looked up:

| Opus status | Sablier status |
| --- | --- |
| `to-do`, `todo`, `open`, `new`, `backlog`, `planned` | `to-do` |
| `in-progress`, `doing`, `started`, `wip`, `active` | `in-progress` |
| `in-review`, `review`, `reviewing`, `testing`, `qa` | `in-review` |
| `done`, `closed`, `complete`, `completed`, `resolved`, `finished` | `done` |

An unknown status keeps the current value on update and becomes `to-do` on create.
The table lives in `modules/antenne/status.go`.

## Registration

The emit and listen lists in `modules/antenne/service.go` (`EmitEvents`, `ListenEvents`) must
equal `antenne.yaml`. A test checks it.

## Cleanup of rows created by replayed legacy events

Before this guard, `project.updated` for an unknown project created it with owner 1 and no
space, and replayed `task.*` events could add tasks under such projects. Locally created rows
always carry a stable id, so a junk row is one whose `facile_id` is not stable, was never used
for time, and was synced (owner 1, no space). Run on a restored copy first, review the SELECT
output, then run the DELETEs in one transaction. Nothing here has been run against production.

```sql
BEGIN;

CREATE TEMP TABLE junk_projects AS
SELECT p.id FROM projects p
WHERE p.facile_id !~ '^(fac_[0-9a-f]{20}|[A-Z2-7]{26}|[0-9a-f]{32})$'
  AND p.owner_id = 1 AND p.space_id IS NULL
  AND NOT EXISTS (SELECT 1 FROM time_entries e WHERE e.project_id = p.id);

CREATE TEMP TABLE junk_tasks AS
SELECT t.id FROM tasks t
WHERE (t.facile_id !~ '^(fac_[0-9a-f]{20}|[A-Z2-7]{26}|[0-9a-f]{32})$' OR t.project_id IN (SELECT id FROM junk_projects))
  AND NOT EXISTS (SELECT 1 FROM time_entries e WHERE e.task_id = t.id);

SELECT 'project' AS kind, p.id, p.name, p.facile_id, p.created_at FROM projects p WHERE p.id IN (SELECT id FROM junk_projects)
UNION ALL
SELECT 'task', t.id, t.name, t.facile_id, t.created_at FROM tasks t WHERE t.id IN (SELECT id FROM junk_tasks);

DELETE FROM tasks WHERE id IN (SELECT id FROM junk_tasks);
DELETE FROM projects WHERE id IN (SELECT id FROM junk_projects)
  AND NOT EXISTS (SELECT 1 FROM tasks WHERE tasks.project_id = projects.id);

COMMIT;
```

Time entries are never touched; any project or task that has one is left for manual review.
Legitimate Opus rows minted in the TS era carry cuid2 ids and match the pattern too, which is
why the SELECT must be read before the DELETE. Roll back instead of committing if a real
project appears. The matching pool ledger rows can stay.
