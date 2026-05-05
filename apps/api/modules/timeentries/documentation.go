package timeentries

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "time-entries",
	Description: "Time tracking routes.",
	Routes: []documentation.Route{
		{Method: "GET", Path: "/time-entries", Summary: "List entries", Auth: "bearer token required"},
		{Method: "GET", Path: "/time-entries/running", Summary: "Get running timer", Auth: "bearer token required"},
		{Method: "POST", Path: "/time-entries/start", Summary: "Start timer", Auth: "bearer token required"},
		{Method: "POST", Path: "/time-entries/stop", Summary: "Stop running timer", Auth: "bearer token required"},
		{Method: "POST", Path: "/time-entries", Summary: "Create manual entry", Auth: "bearer token required"},
		{Method: "PUT", Path: "/time-entries/{id}", Summary: "Update entry", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/time-entries/{id}", Summary: "Delete entry", Auth: "bearer token required"},
	},
}
