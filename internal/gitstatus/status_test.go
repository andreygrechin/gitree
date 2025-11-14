package gitstatus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreygrechin/gitree/internal/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRepoWithState creates a test repository with specific Git state
func createTestRepoWithState(t *testing.T, state string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "gitree-gitstatus-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	repo, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0o644)
	require.NoError(t, err)

	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	sig := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
		When:  time.Now(),
	}

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: sig,
	})
	require.NoError(t, err)

	// Apply state modifications
	switch state {
	case "with-remote":
		// Add a remote
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{"https://github.com/test/repo.git"},
		})
		require.NoError(t, err)

	case "detached":
		// Checkout to detached HEAD
		head, err := repo.Head()
		require.NoError(t, err)
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: head.Hash(),
		})
		require.NoError(t, err)

	case "with-changes":
		// Modify file to create uncommitted changes
		err = os.WriteFile(testFile, []byte("modified content"), 0o644)
		require.NoError(t, err)

	case "with-stash":
		// Create a stash
		// Note: go-git doesn't directly support stash creation, so we'll manually create the ref
		head, err := repo.Head()
		require.NoError(t, err)
		stashRef := plumbing.NewHashReference("refs/stash", head.Hash())
		err = repo.Storer.SetReference(stashRef)
		require.NoError(t, err)

	case "with-ahead":
		// Add a remote first
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{"https://github.com/test/repo.git"},
		})
		require.NoError(t, err)

		// Get the current branch name
		head, err := repo.Head()
		require.NoError(t, err)
		branchName := head.Name().Short()

		// Store the first commit hash
		firstCommitHash := head.Hash()

		// Create another commit to be ahead
		testFile2 := filepath.Join(tempDir, "test2.txt")
		err = os.WriteFile(testFile2, []byte("new file"), 0o644)
		require.NoError(t, err)
		_, err = worktree.Add("test2.txt")
		require.NoError(t, err)
		_, err = worktree.Commit("Second commit", &git.CommitOptions{
			Author: sig,
		})
		require.NoError(t, err)

		// Set remote tracking ref to first commit (so we're 1 ahead)
		remoteBranchName := fmt.Sprintf("refs/remotes/origin/%s", branchName)
		remoteRef := plumbing.NewHashReference(plumbing.ReferenceName(remoteBranchName), firstCommitHash)
		err = repo.Storer.SetReference(remoteRef)
		require.NoError(t, err)

	case "bare":
		// Close current repo and create bare repo
		tempDirBare, err := os.MkdirTemp("", "gitree-bare-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tempDirBare) })

		_, err = git.PlainInit(tempDirBare, true)
		require.NoError(t, err)

		return tempDirBare
	}

	return tempDir
}

// T032: Test helper to initialize test repos with known states
func TestCreateTestRepoHelper(t *testing.T) {
	tests := []struct {
		name  string
		state string
	}{
		{"basic repo", "basic"},
		{"repo with remote", "with-remote"},
		{"detached HEAD", "detached"},
		{"with uncommitted changes", "with-changes"},
		{"with stash", "with-stash"},
		{"bare repository", "bare"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := createTestRepoWithState(t, tt.state)
			assert.DirExists(t, repoPath)

			if tt.state != "bare" {
				assert.DirExists(t, filepath.Join(repoPath, ".git"))
			} else {
				assert.FileExists(t, filepath.Join(repoPath, "HEAD"))
			}
		})
	}
}

// T033: Test Extract() getting branch name
func TestExtract_GetsBranchName(t *testing.T) {
	repoPath := createTestRepoWithState(t, "basic")

	ctx := context.Background()
	status, err := Extract(ctx, repoPath, nil)

	require.NoError(t, err)
	require.NotNil(t, status)
	// Branch can be "main" or "master" depending on Git version
	assert.Contains(t, []string{"main", "master"}, status.Branch)
	assert.False(t, status.IsDetached)
}

// T034: Test Extract() detecting detached HEAD
func TestExtract_DetectsDetachedHEAD(t *testing.T) {
	repoPath := createTestRepoWithState(t, "detached")

	ctx := context.Background()
	status, err := Extract(ctx, repoPath, nil)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "DETACHED", status.Branch)
	assert.True(t, status.IsDetached)
}

// T035: Test Extract() calculating ahead/behind counts
func TestExtract_CalculatesAheadBehindCounts(t *testing.T) {
	repoPath := createTestRepoWithState(t, "with-ahead")

	ctx := context.Background()
	status, err := Extract(ctx, repoPath, nil)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, 1, status.Ahead, "should be 1 commit ahead")
	assert.Equal(t, 0, status.Behind, "should be 0 commits behind")
	assert.True(t, status.HasRemote)
}

// T036: Test Extract() detecting no remote
func TestExtract_DetectsNoRemote(t *testing.T) {
	repoPath := createTestRepoWithState(t, "basic")

	ctx := context.Background()
	status, err := Extract(ctx, repoPath, nil)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.False(t, status.HasRemote)
	assert.Equal(t, 0, status.Ahead)
	assert.Equal(t, 0, status.Behind)
}

// T037: Test Extract() detecting stashes
func TestExtract_DetectsStashes(t *testing.T) {
	repoPath := createTestRepoWithState(t, "with-stash")

	ctx := context.Background()
	status, err := Extract(ctx, repoPath, nil)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.True(t, status.HasStashes)
}

// T038: Test Extract() detecting uncommitted changes
func TestExtract_DetectsUncommittedChanges(t *testing.T) {
	repoPath := createTestRepoWithState(t, "with-changes")

	ctx := context.Background()
	status, err := Extract(ctx, repoPath, nil)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.True(t, status.HasChanges)
}

// T039: Test Extract() handling bare repositories
func TestExtract_HandlesBareRepositories(t *testing.T) {
	repoPath := createTestRepoWithState(t, "bare")

	ctx := context.Background()
	status, err := Extract(ctx, repoPath, nil)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.False(t, status.HasChanges, "bare repo should not have changes")
	// Bare repo may or may not have a branch depending on initialization
}

// T040: Test Extract() handling corrupted repos
func TestExtract_HandlesCorruptedRepos(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitree-corrupt-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a fake .git directory without proper structure
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0o755)
	require.NoError(t, err)

	ctx := context.Background()
	status, err := Extract(ctx, tempDir, nil)

	// Should return error for corrupted repo
	assert.Error(t, err)
	// Status may be nil or partial
	if status != nil {
		assert.NotEmpty(t, status.Error)
	}
}

// T041: Test Extract() respecting context timeout
func TestExtract_RespectsContextTimeout(t *testing.T) {
	repoPath := createTestRepoWithState(t, "basic")

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a moment to ensure timeout fires
	time.Sleep(10 * time.Millisecond)

	status, err := Extract(ctx, repoPath, nil)

	// Should timeout
	assert.Error(t, err)
	// Status may be partial or nil
	if status != nil {
		assert.NotEmpty(t, status.Error)
	}
}

// T042: Test ExtractBatch() concurrent processing
func TestExtractBatch_ConcurrentProcessing(t *testing.T) {
	// Create multiple test repos
	repos := make(map[string]*models.Repository)
	for i := 0; i < 5; i++ {
		repoPath := createTestRepoWithState(t, "basic")
		repoName := fmt.Sprintf("repo%d", i)
		repos[repoPath] = &models.Repository{
			Path: repoPath,
			Name: repoName,
		}
	}

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:        10 * time.Second,
		MaxConcurrency: 3,
	}

	startTime := time.Now()
	statuses, err := ExtractBatch(ctx, repos, opts)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.Len(t, statuses, 5)

	// Verify all repos have status
	for path, status := range statuses {
		assert.NotNil(t, status, "repo %s should have status", path)
		assert.Contains(t, []string{"main", "master"}, status.Branch)
	}

	// Concurrent processing should be faster than sequential (rough check)
	// If sequential, 5 repos would take at least 5x single repo time
	// With concurrency, should be much faster
	t.Logf("Batch processing took: %v", duration)
	assert.Less(t, duration, 5*time.Second, "should complete quickly with concurrency")
}

// Additional test: Extract with custom timeout option
func TestExtract_WithTimeoutOption(t *testing.T) {
	repoPath := createTestRepoWithState(t, "basic")

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout: 5 * time.Second,
	}

	status, err := Extract(ctx, repoPath, opts)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Contains(t, []string{"main", "master"}, status.Branch)
}

// Additional test: Extract handles non-existent path
func TestExtract_NonExistentPath(t *testing.T) {
	ctx := context.Background()
	status, err := Extract(ctx, "/nonexistent/path", nil)

	assert.Error(t, err)
	if status != nil {
		assert.NotEmpty(t, status.Error)
	}
}

// Test ExtractBatch with empty repos map
func TestExtractBatch_EmptyRepos(t *testing.T) {
	ctx := context.Background()
	repos := make(map[string]*models.Repository)

	statuses, err := ExtractBatch(ctx, repos, nil)

	require.NoError(t, err)
	assert.Empty(t, statuses)
}

// Test ExtractBatch respects context cancellation
func TestExtractBatch_RespectsContextCancellation(t *testing.T) {
	// Create multiple test repos
	repos := make(map[string]*models.Repository)
	for i := 0; i < 10; i++ {
		repoPath := createTestRepoWithState(t, "basic")
		repos[repoPath] = &models.Repository{
			Path: repoPath,
			Name: fmt.Sprintf("repo%d", i),
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	statuses, err := ExtractBatch(ctx, repos, nil)

	// Should handle cancellation gracefully
	// May return partial results or error
	_ = statuses
	_ = err
	// The key is that it shouldn't hang
}
