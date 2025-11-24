package models

// BatchResult represents the result of a batch Git status extraction operation.
type BatchResult struct {
	Statuses     map[string]*GitStatus
	FailedRepos  []string
	SuccessCount int
	FailureCount int
}
