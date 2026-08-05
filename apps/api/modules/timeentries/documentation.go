package timeentries

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "time-entries",
	Description: "Time tracking routes.",
	Routes: []documentation.Route{
		{Method: "GET", Path: "/api/time-entries", Summary: "List entries", Auth: "bearer token required"},
		{Method: "GET", Path: "/api/time-entries/running", Summary: "Get running timer", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/time-entries/start", Summary: "Start timer", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/time-entries/stop", Summary: "Stop running timer", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/time-entries", Summary: "Create manual entry", Auth: "bearer token required"},
		{Method: "PUT", Path: "/api/time-entries/{id}", Summary: "Update entry", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/api/time-entries/{id}", Summary: "Delete entry", Auth: "bearer token required"},
	},
}
