package auth

import (
	"context"
	"database/sql"
	"strconv"

	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const (
	personalProjects = `SELECT id FROM projects WHERE space_id IS NULL AND owner_id = ANY($1)`

	deletePersonalEntries  = `DELETE FROM time_entries WHERE project_id IN (` + personalProjects + `)`
	deletePersonalTasks    = `DELETE FROM tasks WHERE project_id IN (` + personalProjects + `)`
	deletePersonalProjects = `DELETE FROM projects WHERE space_id IS NULL AND owner_id = ANY($1)`
	deleteOwnEntries       = `DELETE FROM time_entries WHERE user_id = ANY($1)`
	anonymiseTaskActor     = `UPDATE tasks SET actor_id = NULL WHERE actor_id = ANY($1)`
	transferSpaceProjects  = `UPDATE projects SET owner_id = COALESCE((
		SELECT m.user_id FROM space_members m JOIN users u ON u.id = m.user_id
		WHERE CAST(m.space_id AS text) = CAST(projects.space_id AS text) AND m.user_id <> ALL($1) AND u.kind <> 'agent'
		ORDER BY CASE WHEN m.role IN ('owner', 'admin') THEN 0 ELSE 1 END, m.joined_at, m.id
		LIMIT 1), 0)
		WHERE space_id IS NOT NULL AND owner_id = ANY($1)`
	deleteHeartbeats       = `DELETE FROM heartbeats WHERE user_id = ANY($1)`
	deleteDismissals       = `DELETE FROM draft_dismissals WHERE user_id = ANY($1)`
	deletePushSubscription = `DELETE FROM push_subscriptions WHERE user_id = ANY($1)`
	deleteMemberships      = `DELETE FROM space_members WHERE user_id = ANY($1)`
	deleteOutboxMentions   = `DELETE FROM antenne_outbox o
		WHERE EXISTS (SELECT 1 FROM users u WHERE u.id = ANY($1) AND coalesce(u.email, '') <> ''
			AND strpos(o.payload::text, '"' || lower(u.email) || '"') > 0)
		   OR EXISTS (SELECT 1 FROM porte_identities i WHERE i.user_id = ANY($1)
			AND strpos(o.payload::text, '"' || i.subject || '"') > 0)`
	deleteAgents = `DELETE FROM users WHERE kind = 'agent' AND owner_user_id = $1`
)

// eraseData is the kit's EraseData: it deletes what an erased person owns, in
// the transaction that then deletes their users row. Their personal projects
// with their tasks and time entries, every time entry they or their agents
// logged, their agent users, heartbeats, dismissed drafts, push subscription,
// memberships and pending MCP actions go, and so does every queued or recently
// sent bus event that names their email or Registre subject. A space project they owned passes
// to the space's longest-standing admin (owner or admin role), else its
// longest-standing member, agents excluded; only a space left with nobody
// keeps owner_id 0. A task they were assigned stays with no assignee.
func eraseData(ctx context.Context, tx *sql.Tx, userID int64) error {
	ids := []int64{userID}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM users WHERE kind = 'agent' AND owner_user_id = $1`, userID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, statement := range []string{
		deleteOutboxMentions, deletePersonalEntries, deletePersonalTasks, deletePersonalProjects, deleteOwnEntries,
		anonymiseTaskActor, transferSpaceProjects, deleteHeartbeats, deleteDismissals,
		deletePushSubscription, deleteMemberships,
	} {
		if _, err := tx.ExecContext(ctx, statement, ids); err != nil {
			return err
		}
	}
	if err := mcpkit.ErasePending(ctx, tx, pendingOwners(ids)...); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, deleteAgents, userID)
	return err
}

func pendingOwners(ids []int64) []string {
	owners := make([]string, len(ids))
	for i, id := range ids {
		owners[i] = strconv.FormatInt(id, 10)
	}
	return owners
}
