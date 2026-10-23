# Opus sync

Sablier and Opus keep projects and tasks in step over the Antenne pool. Both
apps emit `project.{created,updated,deleted}` and `task.{created,updated,deleted}`
for local user changes and apply the same events from the other app.

## Rules

- **Identity.** Every project and task has a unique `facile_id` (`fac_` plus hex), minted at
  creation. Startup backfills rows that lack one. Inbound events match rows by `facile_id` only.
- **No loops.** Events whose `app` is `Sablier` are ignored. Applying an inbound event writes
  straight to the database and never queues an outbound event.
- **Idempotency.** Each applied event's `idempotency_key` is stored in the ledger
  (`AntenneProcessedEvent`, pruned after 35 days). A replay does nothing. A dropped event is
  not recorded, so it can be replayed.
- **Tasks need their project.** A task event whose `project_facile_id` matches no local project
  is dropped with a warning. No orphan task is created.
- **Updates for unknown rows** create the row (projects and tasks), except tasks without a known project.
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
