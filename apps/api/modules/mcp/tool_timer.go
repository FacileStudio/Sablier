package mcp

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusStopped    = "stopped"
	statusStarted    = "started"
	statusNotRunning = "not_running"
)

// status is sablier_timer_status.
func (t *toolset) status(ctx context.Context, req *sdk.CallToolRequest, in statusInput) (*sdk.CallToolResult, timerOutput, error) {
	userID, c, err := t.callerClock(ctx, req, in.Timezone)
	if err != nil {
		return nil, timerOutput{}, err
	}
	running, err := t.running(ctx, userID, c)
	if err != nil || running == nil {
		return nil, timerOutput{Status: statusStopped, Message: "No timer is running. Start one with sablier_start_timer."}, err
	}
	return nil, timerOutput{Status: running.State, Entry: running}, nil
}

// start is sablier_start_timer.
func (t *toolset) start(ctx context.Context, req *sdk.CallToolRequest, in startInput) (*sdk.CallToolResult, timerOutput, error) {
	userID, c, err := t.callerClock(ctx, req, in.Timezone)
	if err != nil {
		return nil, timerOutput{}, err
	}
	running, err := t.running(ctx, userID, c)
	if err != nil {
		return nil, timerOutput{}, err
	}
	if running != nil {
		return nil, timerOutput{}, fmt.Errorf("a timer is already running on %s / %s since %s. Call sablier_stop_timer "+
			"first, then start this one", running.Project, running.Task, running.StartedAt)
	}
	taskID, err := t.taskFor(ctx, userID, in.ProjectID, in.TaskID, in.Task)
	if err != nil {
		return nil, timerOutput{}, err
	}
	entry, err := t.Entries.Start(ctx, userID, in.ProjectID, taskID)
	if err != nil {
		return nil, timerOutput{}, mcpkit.Failure(err, hintProject)
	}
	out, err := t.entryView(ctx, userID, entry.ID, c)
	return nil, timerOutput{Status: statusStarted, Entry: out}, err
}

// stop is sablier_stop_timer. With no timer running it changes nothing and says so.
func (t *toolset) stop(ctx context.Context, req *sdk.CallToolRequest, in statusInput) (*sdk.CallToolResult, timerOutput, error) {
	userID, c, err := t.callerClock(ctx, req, in.Timezone)
	if err != nil {
		return nil, timerOutput{}, err
	}
	entry, err := t.Entries.Stop(ctx, userID)
	var known *errors.Error
	if stderrors.As(err, &known) && known.Code == "not_found" {
		return nil, timerOutput{Status: statusNotRunning, Message: "No timer was running, so nothing changed."}, nil
	}
	if err != nil {
		return nil, timerOutput{}, mcpkit.Failure(err, "")
	}
	out, err := t.entryView(ctx, userID, entry.ID, c)
	return nil, timerOutput{Status: statusStopped, Entry: out}, err
}

// callerClock reads the caller and the call's timezone.
func (t *toolset) callerClock(ctx context.Context, req *sdk.CallToolRequest, timezone string) (int64, clock, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return 0, clock{}, err
	}
	c, err := clockOf(timezone, time.Now())
	return userID, c, err
}

// running renders the caller's running entry, or nil.
func (t *toolset) running(ctx context.Context, userID int64, c clock) (*entryOut, error) {
	entry, err := t.Entries.Running(ctx, userID)
	if err != nil || entry == nil {
		return nil, mcpkit.Failure(err, "")
	}
	return t.entryView(ctx, userID, entry.ID, c)
}

// entryView loads one of the caller's entries with its names and renders it.
func (t *toolset) entryView(ctx context.Context, userID, entryID int64, c clock) (*entryOut, error) {
	entry, err := t.Entries.Own(ctx, userID, entryID)
	if err != nil {
		return nil, mcpkit.Failure(err, hintEntry)
	}
	out := view{clock: c, callerID: userID}.entry(*entry)
	return &out, nil
}
