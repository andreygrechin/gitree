package gitstatus

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andreygrechin/gitree/internal/models"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
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

	// Fetch enables fetching from origin remote before calculating ahead/behind counts
	Fetch bool

	// FetchRetries is the number of retry attempts for failed fetch operations
	FetchRetries int
}

const (
	defaultExtractTimeout  = 10 * time.Second
	defaultMaxConcurrency  = 10
	defaultFetchRetries    = 3
	maxFilesPerCategory    = 20
	thresholdSlowOperation = 100 * time.Millisecond
)

var (
	errNoRemotes                = errors.New("no remotes configured")
	errGitignorePathNotAbsolute = errors.New("gitignore path must be absolute")
)

// checkContext returns an error if the context has been canceled.
func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// debugPrintf formats the message using fmt.Sprintf, adds a "DEBUG: " prefix, and outputs it to stderr.
func debugPrintf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "DEBUG: %s\n", message)
}

// DefaultOptions returns sensible default options.
func DefaultOptions() *ExtractOptions {
	return &ExtractOptions{
		Timeout:        defaultExtractTimeout,
		MaxConcurrency: defaultMaxConcurrency,
		Fetch:          true,
		FetchRetries:   defaultFetchRetries,
	}
}

// Extract retrieves Git status information for a single repository.
func Extract(ctx context.Context, repoPath string, opts *ExtractOptions, ignorePatterns []gitignore.Pattern) (*models.GitStatus, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	if opts.Debug {
		debugPrintf("Starting status extraction for repository: %s", repoPath)
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
		status, err := extractGitStatus(ctx, repoPath, opts, ignorePatterns)
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
// It checks for context cancellation between expensive operations to allow early termination.
func extractGitStatus(
	ctx context.Context, repoPath string, opts *ExtractOptions, ignorePatterns []gitignore.Pattern,
) (*models.GitStatus, error) {
	startTime := time.Now()

	// Open repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Check for cancellation after opening repository
	if err := checkContext(ctx); err != nil {
		return nil, err
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

	// Check for cancellation before expensive ahead/behind calculation
	if err := checkContext(ctx); err != nil {
		return nil, err
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

	// Check for cancellation before expensive worktree status scan
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	// Check for uncommitted changes
	if err := extractUncommittedChanges(repo, status, opts, ignorePatterns); err != nil {
		// Non-fatal for bare repos
		if !errors.Is(err, git.ErrIsBareRepository) {
			if status.Error == "" {
				status.Error = err.Error()
			}
		}
	}

	// Debug output
	if opts.Debug {
		printDebugSummary(repoPath, status, startTime)
	}

	return status, nil
}

// printDebugSummary prints debug timing and status summary.
func printDebugSummary(repoPath string, status *models.GitStatus, startTime time.Time) {
	// Timing (only if >100ms)
	duration := time.Since(startTime)
	if duration > thresholdSlowOperation {
		debugPrintf("Repository %s status extraction: %dms", repoPath, duration.Milliseconds())
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

	debugPrintf("Repository %s: %s", repoPath, strings.Join(statusParts, ", "))
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

// readGitignoreFile reads a gitignore file directly and returns patterns.
func readGitignoreFile(path string) ([]gitignore.Pattern, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Verify the path is absolute to prevent relative path manipulation
	if !filepath.IsAbs(cleanPath) {
		return nil, fmt.Errorf("%w: %s", errGitignorePathNotAbsolute, path)
	}

	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close() // Ignore close error on read-only file
	}()

	scanner := bufio.NewScanner(f)
	var patterns []gitignore.Pattern

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
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
			debugPrintf("%s files (%d): %s", category, len(files), strings.Join(files, ", "))
		} else {
			debugPrintf("%s files (%d): %s", category, len(files), strings.Join(files[:maxFilesPerCategory], ", "))
			debugPrintf("...and %d more %s files", len(files)-maxFilesPerCategory, strings.ToLower(category))
		}
	}

	printFileList("Modified", modifiedFiles)
	printFileList("Untracked", untrackedFiles)
	printFileList("Staged", stagedFiles)
	printFileList("Deleted", deletedFiles)
}

// loadGlobalIgnorePatterns loads global gitignore patterns from core.excludesfile and default locations.
func loadGlobalIgnorePatterns(osFS billy.Filesystem, opts *ExtractOptions) ([]gitignore.Pattern, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if opts.Debug {
			debugPrintf("Failed to get home directory: %v", err)
		}

		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	if opts.Debug {
		debugPrintf("User home directory: %s", homeDir)
	}

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(homeDir, ".config")
	}

	debugGlobalIgnoreLocations(osFS, homeDir, xdgConfigHome, opts)

	// Try loading from core.excludesfile first from ~/.gitconfig
	gitconfigPatterns, err := gitignore.LoadGlobalPatterns(osFS)
	if err != nil {
		if opts.Debug {
			debugPrintf("Failed to load global gitignore patterns from core.excludesfile in ~/.gitconfig: %v", err)
		}
	}
	if gitconfigPatterns != nil {
		if opts.Debug {
			debugPrintf("Loaded %d patterns from core.excludesfile", len(gitconfigPatterns))
		}

		return gitconfigPatterns, nil
	}

	if opts.Debug {
		debugPrintf("Failed to load global gitignore patterns from core.excludesfile: %v", err)
	}

	// Fall back to default location
	return loadDefaultGlobalIgnore(xdgConfigHome, opts)
}

// debugGlobalIgnoreLocations prints debug information about gitignore file locations.
func debugGlobalIgnoreLocations(osFS billy.Filesystem, homeDir, xdgConfigHome string, opts *ExtractOptions) {
	if !opts.Debug {
		return
	}

	gitconfigPath := filepath.Join(homeDir, ".gitconfig")
	debugPrintf("Looking for .gitconfig at: %s", gitconfigPath)
	if stat, err := osFS.Stat(gitconfigPath); err == nil {
		debugPrintf(".gitconfig found (size: %d bytes)", stat.Size())
	} else {
		debugPrintf(".gitconfig not found or inaccessible: %v", err)
	}

	defaultIgnorePath := filepath.Join(xdgConfigHome, "git", "ignore")
	debugPrintf("Checking default global ignore at: %s", defaultIgnorePath)
	if stat, err := osFS.Stat(defaultIgnorePath); err == nil {
		debugPrintf("Default global ignore found (size: %d bytes)", stat.Size())
	} else {
		debugPrintf("Default global ignore not found: %v", err)
	}
}

// loadDefaultGlobalIgnore loads patterns from the default global gitignore location.
func loadDefaultGlobalIgnore(xdgConfigHome string, opts *ExtractOptions) ([]gitignore.Pattern, error) {
	defaultIgnorePath := filepath.Join(xdgConfigHome, "git", "ignore")
	defaultPatterns, err := readGitignoreFile(defaultIgnorePath)

	if err == nil && len(defaultPatterns) > 0 {
		if opts.Debug {
			debugPrintf("Loaded %d patterns from default global ignore: %s", len(defaultPatterns), defaultIgnorePath)
		}

		return defaultPatterns, nil
	}

	if opts.Debug && err != nil {
		debugPrintf("Could not read default global ignore: %v", err)
	}

	return nil, err
}

// extractUncommittedChanges checks for uncommitted changes in the working tree.
func extractUncommittedChanges(
	repo *git.Repository,
	status *models.GitStatus,
	opts *ExtractOptions,
	ignorePatterns []gitignore.Pattern,
) error {
	worktree, err := repo.Worktree()
	if err != nil {
		// Likely a bare repository
		status.HasChanges = false
		if errors.Is(err, git.ErrIsBareRepository) {
			return git.ErrIsBareRepository
		}

		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add pre-loaded global gitignore patterns to align with native git behavior
	if len(ignorePatterns) > 0 {
		worktree.Excludes = append(worktree.Excludes, ignorePatterns...)
		if opts.Debug {
			debugPrintf("Total excludes after adding global patterns: %d", len(worktree.Excludes))
		}
	}

	wtStatus, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get worktree status: %w", err)
	}

	status.HasChanges = !wtStatus.IsClean()

	if opts.Debug && status.HasChanges {
		categorizeAndPrintFiles(wtStatus)
	}

	return nil
}

// ExtractBatch extracts Git status for multiple repositories concurrently.
func ExtractBatch(
	ctx context.Context, repos map[string]*models.Repository, opts *ExtractOptions,
) *models.BatchResult {
	if opts == nil {
		opts = DefaultOptions()
	}

	batchResult := &models.BatchResult{
		Statuses:    make(map[string]*models.GitStatus),
		FailedRepos: []string{},
	}

	if len(repos) == 0 {
		return batchResult
	}

	// Fetch phase (if enabled)
	if opts.Fetch {
		if opts.Debug {
			debugPrintf("Starting fetch phase for %d repositories", len(repos))
		}
		fetchBatch(ctx, repos, opts, batchResult)
		if opts.Debug {
			stats := batchResult.FetchStats
			debugPrintf("Fetch complete: %d attempted, %d successful, %d skipped, %d failed",
				stats.TotalAttempted, stats.Successful, stats.Skipped, stats.Failed)
		}
	}

	// Load global gitignore patterns once for all repositories
	osFS := osfs.New("/")
	ignorePatterns, err := loadGlobalIgnorePatterns(osFS, opts)
	if err != nil && opts.Debug {
		debugPrintf("Failed to load global ignore patterns: %v", err)
	}
	// Use empty slice if loading failed
	if ignorePatterns == nil {
		ignorePatterns = []gitignore.Pattern{}
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
			status, err := Extract(ctx, repoPath, opts, ignorePatterns)
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
	for r := range results {
		if r.status != nil {
			batchResult.Statuses[r.path] = r.status
			if r.status.Error != "" {
				batchResult.FailedRepos = append(batchResult.FailedRepos, r.path)
				batchResult.FailureCount++
			} else {
				batchResult.SuccessCount++
			}
		} else if r.err != nil {
			// Fallback if status is nil but error is present (though Extract should handle this)
			batchResult.FailedRepos = append(batchResult.FailedRepos, r.path)
			batchResult.FailureCount++
		}
	}

	return batchResult
}
