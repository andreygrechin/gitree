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
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRepoWithLocalRemote creates a test repository with a local bare repo as remote.
func createTestRepoWithLocalRemote(t *testing.T) string {
	t.Helper()

	// Create bare repo to act as "remote"
	remotePath := t.TempDir()
	_, err := git.PlainInit(remotePath, true)
	require.NoError(t, err)

	// Create working repo
	repoPath := t.TempDir()
	repo, err := git.PlainInit(repoPath, false)
	require.NoError(t, err)

	// Add origin pointing to local bare repo
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remotePath},
	})
	require.NoError(t, err)

	// Create initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0o600)
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

	// Push to remote
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	require.NoError(t, err)

	return repoPath
}

// T_F001: Test fetchFromOrigin with valid origin remote.
func TestFetchFromOrigin_Success(t *testing.T) {
	repoPath := createTestRepoWithLocalRemote(t)

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:      10 * time.Second,
		FetchRetries: 3,
	}

	result := fetchFromOrigin(ctx, repoPath, opts)

	assert.True(t, result.Success || result.AlreadyUpToDate, "fetch should succeed or be already up-to-date")
	assert.False(t, result.Skipped)
	assert.NoError(t, result.Error)
}

// T_F002: Test fetchFromOrigin with no origin remote (should skip, not error).
func TestFetchFromOrigin_NoOrigin(t *testing.T) {
	// Create repo without origin
	repoPath := createTestRepoWithState(t, "basic")

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:      10 * time.Second,
		FetchRetries: 3,
	}

	result := fetchFromOrigin(ctx, repoPath, opts)

	assert.True(t, result.Skipped)
	assert.False(t, result.Success)
	assert.NoError(t, result.Error)
}

// T_F003: Test fetchFromOrigin already up-to-date.
func TestFetchFromOrigin_AlreadyUpToDate(t *testing.T) {
	repoPath := createTestRepoWithLocalRemote(t)

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:      10 * time.Second,
		FetchRetries: 3,
	}

	// First fetch
	result1 := fetchFromOrigin(ctx, repoPath, opts)
	assert.True(t, result1.Success || result1.AlreadyUpToDate)

	// Second fetch should be already up-to-date
	result2 := fetchFromOrigin(ctx, repoPath, opts)
	assert.True(t, result2.Success)
	assert.True(t, result2.AlreadyUpToDate)
}

// T_F004: Test calculateBackoff returns correct delays.
func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 500 * time.Millisecond},
		{2, 1 * time.Second},
		{3, 2 * time.Second},
		{4, 4 * time.Second},
		{5, 8 * time.Second},
		{6, 10 * time.Second}, // capped at max
		{10, 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			actual := calculateBackoff(tt.attempt)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// T_F005: Test context timeout is respected.
func TestFetchFromOrigin_RespectsTimeout(t *testing.T) {
	repoPath := createTestRepoWithLocalRemote(t)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a moment to ensure timeout fires
	time.Sleep(10 * time.Millisecond)

	opts := &ExtractOptions{
		Timeout:      10 * time.Second,
		FetchRetries: 1,
	}

	result := fetchFromOrigin(ctx, repoPath, opts)

	// Should return context error
	require.Error(t, result.Error)
	assert.False(t, result.Success)
}

// T_F006: Test fetchBatch concurrent processing.
func TestFetchBatch_ConcurrentProcessing(t *testing.T) {
	// Create multiple test repos with local remotes
	repos := make(map[string]*models.Repository)
	for i := range 3 {
		repoPath := createTestRepoWithLocalRemote(t)
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
		FetchRetries:   3,
	}
	batchResult := &models.BatchResult{
		Statuses: make(map[string]*models.GitStatus),
	}

	fetchBatch(ctx, repos, opts, batchResult)

	assert.NotNil(t, batchResult.FetchStats)
	assert.Equal(t, 3, batchResult.FetchStats.TotalAttempted)
	assert.Equal(t, 3, batchResult.FetchStats.Successful)
	assert.Equal(t, 0, batchResult.FetchStats.Failed)
}

// T_F007: Test fetchBatch skips bare repositories.
func TestFetchBatch_SkipsBareRepos(t *testing.T) {
	// Create a bare repo and a regular repo
	bareRepoPath := createTestRepoWithState(t, "bare")
	regularRepoPath := createTestRepoWithLocalRemote(t)

	repos := map[string]*models.Repository{
		bareRepoPath:    {Path: bareRepoPath, Name: "bare", IsBare: true},
		regularRepoPath: {Path: regularRepoPath, Name: "regular", IsBare: false},
	}

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:        10 * time.Second,
		MaxConcurrency: 2,
		FetchRetries:   3,
	}
	batchResult := &models.BatchResult{
		Statuses: make(map[string]*models.GitStatus),
	}

	fetchBatch(ctx, repos, opts, batchResult)

	assert.NotNil(t, batchResult.FetchStats)
	// Bare repo should be skipped, regular should be attempted
	assert.Equal(t, 1, batchResult.FetchStats.TotalAttempted)
	assert.Equal(t, 1, batchResult.FetchStats.Skipped) // bare repo
	assert.Equal(t, 1, batchResult.FetchStats.Successful)
}

// T_F008: Test fetchBatch with no origin remote skips gracefully.
func TestFetchBatch_NoOriginSkipped(t *testing.T) {
	// Create repo without origin
	repoPath := createTestRepoWithState(t, "basic")

	repos := map[string]*models.Repository{
		repoPath: {Path: repoPath, Name: "no-origin"},
	}

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:        10 * time.Second,
		MaxConcurrency: 2,
		FetchRetries:   3,
	}
	batchResult := &models.BatchResult{
		Statuses: make(map[string]*models.GitStatus),
	}

	fetchBatch(ctx, repos, opts, batchResult)

	assert.NotNil(t, batchResult.FetchStats)
	// Should be counted as attempted then decremented when skipped
	assert.Equal(t, 0, batchResult.FetchStats.TotalAttempted)
	assert.Equal(t, 1, batchResult.FetchStats.Skipped)
	assert.Equal(t, 0, batchResult.FetchStats.Failed)
}

// T_F009: Test fetchFromOrigin with non-existent path.
func TestFetchFromOrigin_NonExistentPath(t *testing.T) {
	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:      10 * time.Second,
		FetchRetries: 1,
	}

	result := fetchFromOrigin(ctx, "/nonexistent/path", opts)

	assert.False(t, result.Success)
	assert.False(t, result.Skipped)
	assert.Error(t, result.Error)
}

// T_F010: Test ExtractBatch with fetch enabled integrates properly.
func TestExtractBatch_WithFetchEnabled(t *testing.T) {
	// Create repo with local remote
	repoPath := createTestRepoWithLocalRemote(t)

	repos := map[string]*models.Repository{
		repoPath: {Path: repoPath, Name: "test-repo"},
	}

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:        10 * time.Second,
		MaxConcurrency: 2,
		Fetch:          true,
		FetchRetries:   3,
	}

	batchResult := ExtractBatch(ctx, repos, opts)

	// Should have both status and fetch stats
	assert.Len(t, batchResult.Statuses, 1)
	assert.NotNil(t, batchResult.FetchStats)
	assert.Equal(t, 1, batchResult.FetchStats.TotalAttempted)
}

// T_F011: Test ExtractBatch with fetch disabled.
func TestExtractBatch_WithFetchDisabled(t *testing.T) {
	repoPath := createTestRepoWithState(t, "basic")

	repos := map[string]*models.Repository{
		repoPath: {Path: repoPath, Name: "test-repo"},
	}

	ctx := context.Background()
	opts := &ExtractOptions{
		Timeout:        10 * time.Second,
		MaxConcurrency: 2,
		Fetch:          false,
	}

	batchResult := ExtractBatch(ctx, repos, opts)

	// Should have status but no fetch stats
	assert.Len(t, batchResult.Statuses, 1)
	assert.Nil(t, batchResult.FetchStats)
}

// T_F012: Test DefaultOptions sets correct fetch defaults.
func TestDefaultOptions_FetchDefaults(t *testing.T) {
	opts := DefaultOptions()

	assert.True(t, opts.Fetch, "Fetch should be enabled by default")
	assert.Equal(t, defaultFetchRetries, opts.FetchRetries, "FetchRetries should match default")
}
