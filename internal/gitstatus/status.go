package gitstatus

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/andreygrechin/gitree/internal/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ExtractOptions configures the Git status extraction behavior.
type ExtractOptions struct {
	// Timeout is the maximum time to spend extracting status for a single repository
	Timeout time.Duration

	// MaxConcurrency limits the number of repositories processed concurrently in ExtractBatch
	MaxConcurrency int

	// Debug enables debug output for status extraction operations
	Debug bool
}

const (
	defaultExtractTimeout  = 10 * time.Second
	defaultMaxConcurrency  = 10
	maxFilesPerCategory    = 20
	thresholdSlowOperation = 100 * time.Millisecond
)

// DefaultOptions returns sensible default options.
func DefaultOptions() *ExtractOptions {
	return &ExtractOptions{
		Timeout:        defaultExtractTimeout,
		MaxConcurrency: defaultMaxConcurrency,
	}
}

// Extract retrieves Git status information for a single repository.
func Extract(ctx context.Context, repoPath string, opts *ExtractOptions) (*models.GitStatus, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Apply timeout if specified
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Channel to receive result
	resultChan := make(chan *models.GitStatus, 1)
	errorChan := make(chan error, 1)

	go func() {
		status, err := extractGitStatus(repoPath, opts)
		if err != nil {
			errorChan <- err

			return
		}
		resultChan <- status
	}()

	// Wait for result or context cancellation
	select {
	case status := <-resultChan:
		return status, nil
	case err := <-errorChan:
		// Return partial status with error
		partialStatus := &models.GitStatus{
			Branch: "N/A",
			Error:  err.Error(),
		}

		return partialStatus, err
	case <-ctx.Done():
		// Timeout or cancellation
		partialStatus := &models.GitStatus{
			Branch: "N/A",
			Error:  "timeout",
		}

		return partialStatus, ctx.Err()
	}
}

// extractGitStatus performs the actual Git status extraction.
func extractGitStatus(repoPath string, opts *ExtractOptions) (*models.GitStatus, error) {
	startTime := time.Now()

	// Open repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	status := &models.GitStatus{}

	// Extract branch name and detached HEAD status
	if err := extractBranch(repo, status); err != nil {
		status.Branch = "N/A"
		status.Error = err.Error()
	}

	// Check for remote
	if err := extractRemote(repo, status); err != nil {
		// Non-fatal: just means no remote
		status.HasRemote = false
	}

	// Extract ahead/behind counts if remote exists
	if status.HasRemote {
		if err := extractAheadBehind(repo, status); err != nil {
			// Non-fatal: log error but continue
			if status.Error == "" {
				status.Error = err.Error()
			}
		}
	}

	// Check for stashes
	status.HasStashes = extractStashes(repo)

	// Check for uncommitted changes
	if err := extractUncommittedChanges(repo, status, opts); err != nil {
		// Non-fatal for bare repos
		if !errors.Is(err, git.ErrIsBareRepository) {
			if status.Error == "" {
				status.Error = err.Error()
			}
		}
	}

	// Debug output
	if opts != nil && opts.Debug {
		printDebugSummary(repoPath, status, startTime)
	}

	return status, nil
}

// printDebugSummary prints debug timing and status summary.
func printDebugSummary(repoPath string, status *models.GitStatus, startTime time.Time) {
	// Timing (only if >100ms)
	duration := time.Since(startTime)
	if duration > thresholdSlowOperation {
		fmt.Fprintf(os.Stderr, "DEBUG: Repository %s status extraction: %dms\n", repoPath, duration.Milliseconds())
	}

	// Status summary
	statusParts := []string{"branch=" + status.Branch}
	statusParts = append(statusParts, fmt.Sprintf("hasChanges=%t", status.HasChanges))
	if status.HasRemote {
		statusParts = append(statusParts, "hasRemote=true")
		if status.Ahead > 0 {
			statusParts = append(statusParts, fmt.Sprintf("ahead=%d", status.Ahead))
		}
		if status.Behind > 0 {
			statusParts = append(statusParts, fmt.Sprintf("behind=%d", status.Behind))
		}
	} else {
		statusParts = append(statusParts, "hasRemote=false")
	}
	if status.HasStashes {
		statusParts = append(statusParts, "hasStashes=true")
	}

	fmt.Fprintf(os.Stderr, "DEBUG: Repository %s: %s\n", repoPath, strings.Join(statusParts, ", "))
}

// extractBranch extracts the current branch name and detached HEAD status.
func extractBranch(repo *git.Repository, status *models.GitStatus) error {
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
		status.IsDetached = true
		status.Branch = "DETACHED"
	} else {
		status.IsDetached = false
		status.Branch = head.Name().Short()
	}

	return nil
}

var errNoRemotes = errors.New("no remotes configured")

// extractRemote checks if the repository has a remote configured.
func extractRemote(repo *git.Repository, status *models.GitStatus) error {
	remotes, err := repo.Remotes()
	if err != nil {
		return err
	}

	if len(remotes) > 0 {
		status.HasRemote = true

		return nil
	}

	status.HasRemote = false

	return errNoRemotes
}

// extractAheadBehind calculates commits ahead and behind the remote tracking branch.
func extractAheadBehind(repo *git.Repository, status *models.GitStatus) error {
	// Get local HEAD
	head, err := repo.Head()
	if err != nil {
		return err
	}

	// Get remote tracking branch
	branchName := head.Name().Short()
	remoteBranchRefName := plumbing.NewRemoteReferenceName("origin", branchName)

	remoteRef, err := repo.Reference(remoteBranchRefName, false)
	if err != nil {
		// No remote tracking branch
		status.Ahead = 0
		status.Behind = 0

		return err
	}

	// Count commits between local and remote
	localCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	remoteCommit, err := repo.CommitObject(remoteRef.Hash())
	if err != nil {
		return err
	}

	// Count ahead (commits in local not in remote)
	ahead, err := countCommitsBetween(repo, localCommit, remoteCommit)
	if err != nil {
		return err
	}
	status.Ahead = ahead

	// Count behind (commits in remote not in local)
	behind, err := countCommitsBetween(repo, remoteCommit, localCommit)
	if err != nil {
		return err
	}
	status.Behind = behind

	return nil
}

// countCommitsBetween counts commits from 'from' that are not in 'to'.
func countCommitsBetween(repo *git.Repository, from, to *object.Commit) (int, error) {
	// Get all commits reachable from 'to'
	toCommits := make(map[plumbing.Hash]bool)
	iter, err := repo.Log(&git.LogOptions{From: to.Hash})
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	err = iter.ForEach(func(c *object.Commit) error {
		toCommits[c.Hash] = true

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to iterate over commits: %w", err)
	}

	// Count commits reachable from 'from' that are not in 'to'
	count := 0
	iter, err = repo.Log(&git.LogOptions{From: from.Hash})
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	err = iter.ForEach(func(c *object.Commit) error {
		if !toCommits[c.Hash] {
			count++
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// extractStashes checks if the repository has any stashed changes.
func extractStashes(repo *git.Repository) bool {
	stashRef, err := repo.Reference("refs/stash", false)
	if err != nil {
		return false
	}

	return stashRef != nil
}

// categorizeAndPrintFiles categorizes and prints files with truncation.
func categorizeAndPrintFiles(wtStatus git.Status) {
	modifiedFiles := []string{}
	untrackedFiles := []string{}
	stagedFiles := []string{}
	deletedFiles := []string{}

	for filename, fileStatus := range wtStatus {
		switch {
		case fileStatus.Worktree == git.Modified && fileStatus.Staging == git.Unmodified:
			modifiedFiles = append(modifiedFiles, filename)
		case fileStatus.Staging == git.Untracked && fileStatus.Worktree == git.Untracked:
			untrackedFiles = append(untrackedFiles, filename)
		case fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked:
			stagedFiles = append(stagedFiles, filename)
		case fileStatus.Worktree == git.Deleted:
			deletedFiles = append(deletedFiles, filename)
		}
	}

	printFileList := func(category string, files []string) {
		if len(files) == 0 {
			return
		}
		if len(files) <= maxFilesPerCategory {
			fmt.Fprintf(os.Stderr, "DEBUG: %s files (%d): %s\n", category, len(files), strings.Join(files, ", "))
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: %s files (%d): %s\n", category, len(files), strings.Join(files[:maxFilesPerCategory], ", "))
			fmt.Fprintf(os.Stderr, "DEBUG: ...and %d more %s files\n", len(files)-maxFilesPerCategory, strings.ToLower(category))
		}
	}

	printFileList("Modified", modifiedFiles)
	printFileList("Untracked", untrackedFiles)
	printFileList("Staged", stagedFiles)
	printFileList("Deleted", deletedFiles)
}

// extractUncommittedChanges checks for uncommitted changes in the working tree.
func extractUncommittedChanges(repo *git.Repository, status *models.GitStatus, opts *ExtractOptions) error {
	worktree, err := repo.Worktree()
	if err != nil {
		// Likely a bare repository
		status.HasChanges = false

		return err
	}

	// Load global gitignore patterns to align with native git behavior
	globalPatterns, err := gitignore.LoadGlobalPatterns(worktree.Filesystem)
	if err == nil {
		worktree.Excludes = append(worktree.Excludes, globalPatterns...)
	}

	// Load system gitignore patterns
	systemPatterns, err := gitignore.LoadSystemPatterns(worktree.Filesystem)
	if err == nil {
		worktree.Excludes = append(worktree.Excludes, systemPatterns...)
	}

	wtStatus, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get worktree status: %w", err)
	}

	status.HasChanges = !wtStatus.IsClean()

	// Debug file listing
	if opts != nil && opts.Debug && status.HasChanges {
		categorizeAndPrintFiles(wtStatus)
	}

	return nil
}

// ExtractBatch extracts Git status for multiple repositories concurrently.
func ExtractBatch(
	ctx context.Context, repos map[string]*models.Repository, opts *ExtractOptions) (map[string]*models.GitStatus, error,
) {
	if opts == nil {
		opts = DefaultOptions()
	}

	if len(repos) == 0 {
		return make(map[string]*models.GitStatus), nil
	}

	// Create channels
	type result struct {
		path   string
		status *models.GitStatus
		err    error
	}

	results := make(chan result, len(repos))
	semaphore := make(chan struct{}, opts.MaxConcurrency)

	var wg sync.WaitGroup

	// Launch workers for each repository
	for path := range repos {
		wg.Add(1)

		go func(repoPath string) {
			defer wg.Done()

			// Check context before starting
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			// Extract status
			status, err := Extract(ctx, repoPath, opts)
			results <- result{
				path:   repoPath,
				status: status,
				err:    err,
			}
		}(path)
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	statuses := make(map[string]*models.GitStatus)
	for r := range results {
		if r.status != nil {
			statuses[r.path] = r.status
		}
	}

	return statuses, nil
}
