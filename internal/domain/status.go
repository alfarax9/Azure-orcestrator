package domain

// Status is the one name for a worker's condition, used the same way on cards, titles,
// toasts and notifications (DESIGN_SYSTEM.md section 10).
type Status string

// The status vocabulary.
const (
	StatusQueued          Status = "queued"
	StatusWorking         Status = "working"
	StatusNeedsYou        Status = "needs_you"
	StatusPreviewLive     Status = "preview_live"
	StatusWaitingOnClient Status = "waiting_on_client"
	StatusClientFeedback  Status = "client_feedback"
	StatusGateFailed      Status = "gate_failed"
	StatusReadyToShip     Status = "ready_to_ship"
	StatusShipping        Status = "shipping"
	StatusLive            Status = "live"
	StatusRolledBack      Status = "rolled_back"
)

// Label is the English label people read.
func (s Status) Label() string {
	switch s {
	case StatusQueued:
		return "Queued"
	case StatusWorking:
		return "Working"
	case StatusNeedsYou:
		return "Needs you"
	case StatusPreviewLive:
		return "Preview live"
	case StatusWaitingOnClient:
		return "Waiting on client"
	case StatusClientFeedback:
		return "Client feedback"
	case StatusGateFailed:
		return "Gate failed"
	case StatusReadyToShip:
		return "Ready to ship"
	case StatusShipping:
		return "Shipping"
	case StatusLive:
		return "Live"
	case StatusRolledBack:
		return "Rolled back"
	}
	return string(s)
}

// NeedsAttention reports whether the worker waits on the developer: these are the statuses
// of the Board's Needs you column, and the ones that justify a notification (F4-04).
func (s Status) NeedsAttention() bool {
	switch s {
	case StatusNeedsYou, StatusClientFeedback, StatusGateFailed, StatusRolledBack:
		return true
	}
	return false
}

// Status derives the worker's status from its facts. Later stages win: a deployment
// outranks an old gate result, and a running agent outranks a result it is about to replace.
func (w Worker) Status() Status {
	if w.Queued {
		return StatusQueued // not started: nothing else is true of it yet
	}
	switch w.Deployment.State {
	case DeployRolledBack:
		return StatusRolledBack
	case DeployShipping:
		return StatusShipping
	case DeployWatching, DeployLive:
		return StatusLive
	case DeployNone:
	}
	switch {
	case w.Activity == ActivityNeedsInput:
		return StatusNeedsYou
	case w.Activity == ActivityWorking:
		return StatusWorking
	case w.Gate.State == GateFailed:
		return StatusGateFailed
	case w.Gate.State == GatePassed:
		return StatusReadyToShip
	case w.NewFeedback > 0:
		return StatusClientFeedback
	case w.Shared:
		return StatusWaitingOnClient
	case w.PreviewLive:
		return StatusPreviewLive
	}
	// The agent finished its turn and nothing runs yet: it waits for the next message.
	return StatusNeedsYou
}
