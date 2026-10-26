package domain

// Stage is one step of the loop strip every worker walks (DESIGN_SYSTEM.md section 8.9).
type Stage string

// The seven stages, in order.
const (
	StagePrompt   Stage = "Prompt"
	StageWorktree Stage = "Worktree"
	StagePreview  Stage = "Preview"
	StageClient   Stage = "Client"
	StageGate     Stage = "Gate"
	StageLive     Stage = "Live"
	StageWatch    Stage = "Watch"
)

// Stages lists the stages in the order the loop strip shows them.
func Stages() []Stage {
	return []Stage{StagePrompt, StageWorktree, StagePreview, StageClient, StageGate, StageLive, StageWatch}
}

// StageState is how one stage looks.
type StageState string

// Stage states.
const (
	StageDone    StageState = "done"
	StageRunning StageState = "running"
	StagePending StageState = "pending"
	StageFailed  StageState = "failed"
	StageSkipped StageState = "skipped" // for example Client in an internal project
)

type stagePosition struct {
	index int
	state StageState
}

// StageStates derives the state of each of the seven stages, in order, from the facts.
func (w Worker) StageStates() []StageState {
	all := Stages()
	at := func(s Stage, state StageState) stagePosition {
		for i, x := range all {
			if x == s {
				return stagePosition{i, state}
			}
		}
		return stagePosition{}
	}
	pos := w.stagePosition(at)
	client := at(StageClient, StageDone).index
	states := make([]StageState, len(all))
	for i := range all {
		switch {
		case i == client && !w.ClientReview:
			states[i] = StageSkipped
		case i < pos.index:
			states[i] = StageDone
		case i == pos.index:
			states[i] = pos.state
		default:
			states[i] = StagePending
		}
	}
	return states
}

// stagePosition finds the stage the worker is at and how that stage looks. A live
// deployment is past the last stage.
func (w Worker) stagePosition(at func(Stage, StageState) stagePosition) stagePosition {
	switch w.Deployment.State {
	case DeployRolledBack:
		return at(StageWatch, StageFailed)
	case DeployWatching:
		return at(StageWatch, StageRunning)
	case DeployLive:
		return stagePosition{index: len(Stages()), state: StageDone}
	case DeployShipping:
		return at(StageLive, StageRunning)
	case DeployNone:
	}
	switch {
	case w.Gate.State == GateFailed:
		return at(StageGate, StageFailed)
	case w.Gate.State == GatePassed:
		return at(StageLive, StagePending)
	case w.ClientReview && (w.Shared || w.NewFeedback > 0):
		return at(StageClient, StageRunning)
	case w.PreviewLive:
		return at(StagePreview, StageRunning)
	}
	return at(StageWorktree, StageRunning)
}
