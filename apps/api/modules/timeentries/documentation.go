package timeentries

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "time-entries",
	Description: "Time tracking routes.",
	Routes: []documentation.Route{
		{
			Method:  "GET",
			Path:    "/time-entries",
			Summary: "List entries",
			Auth:    "bearer token required",
			QueryParams: []documentation.Field{
				{Name: "space_id", Type: "string", Description: "Filter by space ID"},
				{Name: "running", Type: "boolean", Description: "Filter by running state"},
				{Name: "project_id", Type: "integer", Description: "Filter by project ID"},
				{Name: "user_id", Type: "integer", Description: "Filter by user ID"},
			},
			ResponseBody: ListEntriesResponse{},
		},
		{
			Method:       "GET",
			Path:         "/time-entries/running",
			Summary:      "Get running timer",
			Auth:         "bearer token required",
			ResponseBody: RunningEntryResponse{},
		},
		{
			Method:       "POST",
			Path:         "/time-entries/start",
			Summary:      "Start timer",
			Auth:         "bearer token required",
			RequestBody:  StartTimerRequest{},
			ResponseBody: TimeEntryResponse{},
		},
		{
			Method:       "POST",
			Path:         "/time-entries/stop",
			Summary:      "Stop running timer",
			Auth:         "bearer token required",
			ResponseBody: TimeEntryResponse{},
		},
		{
			Method:       "POST",
			Path:         "/time-entries/pause",
			Summary:      "Pause running timer",
			Auth:         "bearer token required",
			ResponseBody: TimeEntryResponse{},
		},
		{
			Method:       "POST",
			Path:         "/time-entries/resume",
			Summary:      "Resume paused timer",
			Auth:         "bearer token required",
			ResponseBody: TimeEntryResponse{},
		},
		{
			Method:       "POST",
			Path:         "/time-entries",
			Summary:      "Create manual entry",
			Auth:         "bearer token required",
			RequestBody:  CreateEntryRequest{},
			ResponseBody: TimeEntryResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/time-entries/{id}",
			Summary:      "Update entry",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "id", Type: "integer", Description: "Time entry ID"}},
			RequestBody:  UpdateEntryRequest{},
			ResponseBody: TimeEntryResponse{},
		},
		{
			Method:       "DELETE",
			Path:         "/time-entries/{id}",
			Summary:      "Delete entry",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "id", Type: "integer", Description: "Time entry ID"}},
			ResponseBody: DeleteEntryResponse{},
		},
	},
}
