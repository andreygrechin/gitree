package gitstatus

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/andreygrechin/gitree/internal/models"
	"github.com/go-git/go-git/v5"
)

const (
	originRemote       = "origin"
	baseBackoffDelay   = 500 * time.Millisecond
	maxBackoffDelay    = 10 * time.Second
	backoffExponentTwo = 2 // Base for exponential backoff calculation.
)

// FetchResult represents the result of a fetch operation.
type FetchResult struct {
	Success bool
	Skipped bool
	Error   error
	Retries int
}

// fetchFromOrigin fetches from the origin remote with retry logic.
// Returns FetchResult indicating success, skip (no origin), or failure with error.
func fetchFromOrigin(ctx context.Context, repoPath string, opts *ExtractOptions) *FetchResult {
	result := &FetchResult{}

	// Open repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open repository: %w", err)

		return result
	}

	// Check if origin remote exists
	remote, err := repo.Remote(originRemote)
	if err != nil {
		// No origin remote - skip fetch, not an error
		result.Skipped = true

		return result
	}

	// Verify remote has URLs configured
	remoteConfig := remote.Config()
	if len(remoteConfig.URLs) == 0 {
		result.Skipped = true

		return result
	}

	remoteURL := remoteConfig.URLs[0]

	// Perform fetch with retries
	maxRetries := opts.FetchRetries
	if maxRetries <= 0 {
		maxRetries = defaultFetchRetries
	}

	var lastErr error
	for attempt := range maxRetries {
		result.Retries = attempt

		// Check context before each attempt
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()

			return result
		default:
		}

		// Calculate backoff delay for retries
		if attempt > 0 {
			delay := calculateBackoff(attempt)
			if opts.Debug {
				debugPrintf("Retry %d for %s after %v", attempt, repoPath, delay)
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				result.Error = ctx.Err()

				return result
			}
		}

		// Perform fetch with timeout
		fetchErr := performFetch(ctx, repo, remoteURL, opts)
		if fetchErr == nil {
			result.Success = true

			return result
		}

		// Check for "already up-to-date" - this is success, not error
		if errors.Is(fetchErr, git.NoErrAlreadyUpToDate) {
			result.Success = true

			return result
		}

		lastErr = fetchErr

		// Don't retry on certain errors (context canceled, etc.)
		if errors.Is(fetchErr, context.Canceled) || errors.Is(fetchErr, context.DeadlineExceeded) {
			break
		}

		if opts.Debug {
			debugPrintf("Fetch attempt %d failed for %s: %v", attempt+1, repoPath, fetchErr)
		}
	}

	result.Error = fmt.Errorf("fetch failed after retries: %w", lastErr)

	return result
}

// performFetch executes a single fetch operation with timeout.
func performFetch(ctx context.Context, repo *git.Repository, remoteURL string, opts *ExtractOptions) error {
	fetchCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		fetchCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Get authentication for HTTPS URLs
	auth := getAuthForURL(fetchCtx, remoteURL, opts.Debug)

	return repo.FetchContext(fetchCtx, &git.FetchOptions{
		RemoteName: originRemote,
		Auth:       auth,
	})
}

// calculateBackoff returns the backoff delay for the given retry attempt.
// Uses exponential backoff: base * 2^(attempt-1), capped at maxBackoffDelay.
func calculateBackoff(attempt int) time.Duration {
	delay := baseBackoffDelay * time.Duration(math.Pow(backoffExponentTwo, float64(attempt-1)))

	return min(delay, maxBackoffDelay)
}

// fetchBatch performs concurrent fetch operations for multiple repositories.
// Thread-safety note: All FetchStats modifications occur in the main goroutine.
// The spawned goroutines only communicate via the results channel, so no mutex is needed.
func fetchBatch(
	ctx context.Context,
	repos map[string]*models.Repository,
	opts *ExtractOptions,
	batchResult *models.BatchResult,
) {
	if batchResult.FetchStats == nil {
		batchResult.FetchStats = &models.FetchStats{}
	}

	type fetchResultPair struct {
		path   string
		result *FetchResult
	}

	results := make(chan fetchResultPair, len(repos))
	semaphore := make(chan struct{}, opts.MaxConcurrency)

	var wg sync.WaitGroup

	for path, repo := range repos {
		// Skip nil repositories
		if repo == nil {
			continue
		}
		// Skip bare repositories - they typically don't have working trees to fetch into
		if repo.IsBare {
			batchResult.FetchStats.Skipped++

			continue
		}

		wg.Add(1)
		batchResult.FetchStats.TotalAttempted++

		go func(repoPath string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			fetchResult := fetchFromOrigin(ctx, repoPath, opts)
			results <- fetchResultPair{path: repoPath, result: fetchResult}
		}(path)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect fetch results (runs sequentially after all goroutines are spawned)
	for r := range results {
		switch {
		case r.result.Skipped:
			batchResult.FetchStats.Skipped++
			batchResult.FetchStats.TotalAttempted-- // Was counted but skipped
		case r.result.Success:
			batchResult.FetchStats.Successful++
		default:
			batchResult.FetchStats.Failed++
			batchResult.FetchStats.FailedRepos = append(batchResult.FetchStats.FailedRepos, r.path)

			// Store fetch error in the repository's GitStatus
			if repo, exists := repos[r.path]; exists && r.result.Error != nil {
				if repo.GitStatus == nil {
					repo.GitStatus = &models.GitStatus{}
				}
				repo.GitStatus.FetchError = r.result.Error.Error()
			}
		}
	}
}
