package domain

import "time"

// AgentKind is the agent CLI a worker runs.
type AgentKind string

// Supported agents (PRD section 2).
const (
	AgentClaude AgentKind = "claude"
	AgentCodex  AgentKind = "codex"
)

// Activity is what the agent is doing, as last observed from its terminal and transcript.
type Activity string

// Agent activities.
const (
	ActivityIdle       Activity = "idle"        // finished its turn, nothing running
	ActivityWorking    Activity = "working"     // writing or running commands
	ActivityNeedsInput Activity = "needs_input" // waiting for an answer or an approval
)

// GateState is the outcome of the last Ship Gate run.
type GateState string

// Gate states.
const (
	GateNone   GateState = "none"
	GateFailed GateState = "failed"
	GatePassed GateState = "passed"
)

// GateFact is the gate result a worker last had.
type GateFact struct {
	State GateState `json:"state"`
	Cause string    `json:"cause,omitzero"` // short reason when the gate failed
}

// DeployState is where the worker's change is in production.
type DeployState string

// Deployment states.
const (
	DeployNone       DeployState = "none"
	DeployShipping   DeployState = "shipping"
	DeployWatching   DeployState = "watching" // live, inside the 15 minute watch window
	DeployLive       DeployState = "live"
	DeployRolledBack DeployState = "rolled_back"
)

// DeploymentFact is the last deployment of a worker's change.
type DeploymentFact struct {
	State  DeployState `json:"state"`
	Reason string      `json:"reason,omitzero"` // why it was rolled back
}

// Worker is one task with one agent in one worktree. Everything below the identity fields
// is a fact the daemon observed; the status people see is derived from them (F3-06).
type Worker struct {
	ID        WorkerID  `json:"id"`
	ProjectID ProjectID `json:"projectId"`
	Title     string    `json:"title"`
	Branch    string    `json:"branch"`
	Agent     AgentKind `json:"agent"`
	CreatedAt Timestamp `json:"createdAt"`
	Worktree  string    `json:"-"` // path on the server; never sent to clients

	// Queued means the server was already running as many workers as it may, so this one waits
	// for a free slot (F10-05). Its agent has not started.
	Queued bool `json:"queued"`

	// ClientReview is false for internal projects: the Client stage is skipped.
	ClientReview bool           `json:"clientReview"`
	Activity     Activity       `json:"agentActivity"`
	PreviewLive  bool           `json:"previewLive"`
	Shared       bool           `json:"shared"` // the preview link was given to the client
	NewFeedback  int            `json:"newFeedback"`
	Gate         GateFact       `json:"gate"`
	Deployment   DeploymentFact `json:"deployment"`
	Added        int            `json:"added"`
	Removed      int            `json:"removed"`
}

// NewWorker returns a worker that has just been created: its agent is starting.
func NewWorker(project ProjectID, title string, agent AgentKind, branch string, clientReview bool, now time.Time) Worker {
	return Worker{
		ID: NewWorkerID(), ProjectID: project, Title: title, Agent: agent, Branch: branch,
		CreatedAt: Timestamp{now}, ClientReview: clientReview, Activity: ActivityWorking,
		Gate: GateFact{State: GateNone}, Deployment: DeploymentFact{State: DeployNone},
	}
}
