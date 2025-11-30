package models

// FetchStats tracks fetch operation statistics.
type FetchStats struct {
	TotalAttempted int      // Repos where fetch was attempted
	Successful     int      // Successful fetches (including already up-to-date)
	Skipped        int      // Repos skipped (no origin remote, bare repos, etc.)
	Failed         int      // Repos where fetch failed after retries
	FailedRepos    []string // Paths of repos that failed to fetch
}

// BatchResult represents the result of a batch Git status extraction operation.
type BatchResult struct {
	Statuses     map[string]*GitStatus
	FailedRepos  []string
	SuccessCount int
	FailureCount int
	FetchStats   *FetchStats // Fetch operation statistics (nil if fetch disabled)
}
