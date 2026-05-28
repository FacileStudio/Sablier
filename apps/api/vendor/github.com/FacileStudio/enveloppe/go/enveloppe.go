package enveloppe

type App string

const (
	AppSablier  App = "Sablier"
	AppOpus     App = "Opus"
	AppArdoise  App = "Ardoise"
	AppPlume    App = "Plume"
	AppGlouton  App = "Glouton"
	AppVision   App = "Vision"
)

type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)

type ObjectType string

const (
	ObjectProject   ObjectType = "project"
	ObjectTask      ObjectType = "task"
	ObjectUser      ObjectType = "user"
	ObjectTimeEntry ObjectType = "time_entry"
	ObjectInvoice   ObjectType = "invoice"
	ObjectDocument  ObjectType = "document"
)

type Event[T any] struct {
	App            App        `json:"app"`
	Object         ObjectType `json:"object"`
	Action         Action     `json:"action"`
	FacileID       string     `json:"facile_id"`
	Payload        T          `json:"payload"`
	Timestamp      string     `json:"timestamp"`
	IdempotencyKey string     `json:"idempotency_key"`
}

type Project struct {
	FacileID    string  `json:"facile_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

type Task struct {
	FacileID        string `json:"facile_id"`
	ProjectFacileID string `json:"project_facile_id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
}

type User struct {
	FacileID string `json:"facile_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

type TimeEntry struct {
	FacileID     string  `json:"facile_id"`
	TaskFacileID string  `json:"task_facile_id"`
	UserFacileID string  `json:"user_facile_id"`
	StartedAt    string  `json:"started_at"`
	StoppedAt    *string `json:"stopped_at"`
}

type Invoice struct {
	FacileID   string  `json:"facile_id"`
	ClientName string  `json:"client_name"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
}

type Document struct {
	FacileID    string `json:"facile_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	SignerEmail string `json:"signer_email"`
}
