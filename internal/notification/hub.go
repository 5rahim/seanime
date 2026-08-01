package notification

type Hub struct {
}

type Severity string

const (
	SeverityNormal  Severity = "normal"
	SeverityInfo    Severity = "info"
	SeveritySuccess Severity = "success"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Urgency string

const (
	UrgencySilent Urgency = "silent"
	UrgencyNormal Urgency = "normal"
	UrgencyUrgent Urgency = "urgent"
)

type Progress struct {
	Current      int
	Total        int
	Intermediate bool
}

type Notification struct {
	ID        string   `json:"id"`
	EmitterId string   `json:"emitterId"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Severity  Severity `json:"severity"`
	Urgency   Urgency  `json:"urgency"`

	ShouldNotifySystem bool `json:"shouldNotifySystem"`
	HasNotifiedSystem  bool `json:"hasNotifiedSystem"`

	Progress Progress `json:"progress"`
}
