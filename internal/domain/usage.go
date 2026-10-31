package domain

// Usage is the tokens one agent message used, as the agent's own transcript reports it
// (F14-04). Cache reads and writes are kept apart from input because they are priced
// differently, and Input never includes the cache.
type Usage struct {
	// Key names the message, so a message seen twice (a transcript read again after a restart,
	// or one message written as several entries) is counted once.
	Key              string `json:"key"`
	Model            string `json:"model,omitzero"`
	InputTokens      int    `json:"inputTokens"`
	OutputTokens     int    `json:"outputTokens"`
	CacheReadTokens  int    `json:"cacheReadTokens"`
	CacheWriteTokens int    `json:"cacheWriteTokens"`
}

// Empty reports whether the message used nothing, such as a synthetic note the agent wrote.
func (u Usage) Empty() bool {
	return u.InputTokens == 0 && u.OutputTokens == 0 && u.CacheReadTokens == 0 && u.CacheWriteTokens == 0
}

// UsageDay is the usage of one agent and model on one UTC day.
type UsageDay struct {
	Day              string    `json:"day"` // 2026-10-31
	Agent            AgentKind `json:"agent"`
	Model            string    `json:"model"`
	Messages         int       `json:"messages"`
	InputTokens      int       `json:"inputTokens"`
	OutputTokens     int       `json:"outputTokens"`
	CacheReadTokens  int       `json:"cacheReadTokens"`
	CacheWriteTokens int       `json:"cacheWriteTokens"`
}
