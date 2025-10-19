package dto

// PublicLogEntry describes a single public usage log entry.
type PublicLogEntry struct {
	ID        string `json:"id"`
	Model     string `json:"model"`
	Subject   string `json:"subject"` // anonymized user/token identifier when necessary
	Tokens    int64  `json:"tokens,omitempty"`
	Requests  int64  `json:"requests,omitempty"`
	RPM       int64  `json:"rpm,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

// LeaderboardEntry aggregates usage for a given subject.
type LeaderboardEntry struct {
	Subject string `json:"subject"`
	Score   int64  `json:"score"`
	Tokens  int64  `json:"tokens,omitempty"`
	Requests int64 `json:"requests,omitempty"`
	Rank    int    `json:"rank"`
}

// PublicLogsResponse is a simple wrapper for logs and leaderboard returns.
type PublicLogsResponse struct {
	Logs        []PublicLogEntry  `json:"logs"`
	Leaderboard []LeaderboardEntry `json:"leaderboard,omitempty"`
}
