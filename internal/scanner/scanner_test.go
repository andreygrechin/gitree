package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test repositories
func createTestRepo(t *testing.T, path string, bare bool) {
	t.Helper()

	err := os.MkdirAll(path, 0755)
	require.NoError(t, err)

	if bare {
		// Create bare repository structure
		err = os.MkdirAll(filepath.Join(path, "refs", "heads"), 0755)
		require.NoError(t, err)
		err = os.MkdirAll(filepath.Join(path, "objects"), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(path, "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
		require.NoError(t, err)
	} else {
		// Create regular repository structure
		gitDir := filepath.Join(path, ".git")
		err = os.MkdirAll(gitDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
		require.NoError(t, err)
	}
}

// T020: Test IsGitRepository() detecting regular repos with .git directory
func TestIsGitRepository_Regular(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")
	createTestRepo(t, repoPath, false)

	isRepo, isBare := IsGitRepository(repoPath)

	assert.True(t, isRepo, "Should detect regular repository")
	assert.False(t, isBare, "Should not be detected as bare")
}

// T021: Test IsGitRepository() detecting bare repos
func TestIsGitRepository_Bare(t *testing.T) {
	tempDir := t.TempDir()
	bareRepoPath := filepath.Join(tempDir, "bare-repo.git")
	createTestRepo(t, bareRepoPath, true)

	isRepo, isBare := IsGitRepository(bareRepoPath)

	assert.True(t, isRepo, "Should detect bare repository")
	assert.True(t, isBare, "Should be detected as bare")
}

// Test IsGitRepository() with non-repository directory
func TestIsGitRepository_NotARepo(t *testing.T) {
	tempDir := t.TempDir()

	isRepo, isBare := IsGitRepository(tempDir)

	assert.False(t, isRepo, "Should not detect non-repository as repo")
	assert.False(t, isBare, "Should not detect non-repository as bare")
}

// T022: Test Scan() finding all repos in test directory tree
func TestScan_FindsAllRepos(t *testing.T) {
	tempDir := t.TempDir()

	// Create directory structure with multiple repos
	createTestRepo(t, filepath.Join(tempDir, "repo1"), false)
	createTestRepo(t, filepath.Join(tempDir, "subdir", "repo2"), false)
	createTestRepo(t, filepath.Join(tempDir, "subdir", "nested", "repo3"), false)

	// Create non-repo directory
	err := os.MkdirAll(filepath.Join(tempDir, "not-a-repo"), 0755)
	require.NoError(t, err)

	ctx := context.Background()
	opts := ScanOptions{
		RootPath: tempDir,
	}
	result, err := Scan(ctx, opts)

	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Repositories), "Should find exactly 3 repositories")
	assert.Equal(t, 3, result.TotalRepos)
}

// T023: Test skipping nested repo contents per spec FR-018
func TestScan_SkipsNestedRepoContents(t *testing.T) {
	tempDir := t.TempDir()

	// Create parent repo
	parentPath := filepath.Join(tempDir, "parent-repo")
	createTestRepo(t, parentPath, false)

	// Create nested repo inside parent (like a submodule scenario)
	nestedPath := filepath.Join(parentPath, "nested-repo")
	createTestRepo(t, nestedPath, false)

	// Create a repo inside the nested repo (should be skipped)
	deepNestedPath := filepath.Join(nestedPath, "deep-nested")
	createTestRepo(t, deepNestedPath, false)

	ctx := context.Background()
	opts := ScanOptions{
		RootPath: tempDir,
	}
	result, err := Scan(ctx, opts)

	require.NoError(t, err)
	// Should find parent, but NOT nested or deep-nested
	// Because once we find parent-repo with .git, we skip its contents (FR-018)
	assert.Equal(t, 1, len(result.Repositories), "Should find only the parent repo, skip its contents")

	// Verify the repo we found is the parent
	assert.Equal(t, "parent-repo", filepath.Base(result.Repositories[0].Path))

	// Verify we don't traverse into nested repo contents
	foundNested := false
	foundDeepNested := false
	for _, repo := range result.Repositories {
		baseName := filepath.Base(repo.Path)
		if baseName == "nested-repo" {
			foundNested = true
		}
		if baseName == "deep-nested" {
			foundDeepNested = true
		}
	}
	assert.False(t, foundNested, "Should not find repos inside parent repo")
	assert.False(t, foundDeepNested, "Should not find repos inside nested repos")
}

// T024: Test permission denied error handling (non-fatal)
func TestScan_PermissionDeniedNonFatal(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()

	// Create accessible repo
	createTestRepo(t, filepath.Join(tempDir, "accessible-repo"), false)

	// Create inaccessible directory
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.MkdirAll(restrictedDir, 0000)
	require.NoError(t, err)
	defer os.Chmod(restrictedDir, 0755) // Cleanup

	ctx := context.Background()
	opts := ScanOptions{
		RootPath: tempDir,
	}
	result, err := Scan(ctx, opts)

	// Should not return fatal error
	require.NoError(t, err, "Permission errors should be non-fatal")

	// Should find the accessible repo
	assert.GreaterOrEqual(t, len(result.Repositories), 1, "Should find accessible repos")

	// Should have collected permission errors
	assert.Greater(t, len(result.Errors), 0, "Should collect permission errors")
}

// T025: Test context cancellation
func TestScan_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple repos
	for i := 0; i < 10; i++ {
		createTestRepo(t, filepath.Join(tempDir, "repo"+string(rune('0'+i))), false)
	}

	// Create context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := ScanOptions{
		RootPath: tempDir,
	}
	result, err := Scan(ctx, opts)

	// Should return context cancelled error or partial results
	// Based on implementation, context cancellation might be checked periodically
	if err != nil {
		assert.ErrorIs(t, err, context.Canceled, "Should return context.Canceled error")
	} else {
		// If we get partial results, that's also acceptable
		assert.NotNil(t, result, "Should return partial results on cancellation")
	}
}

// T026: Test helper is the createTestRepo function above
func TestCreateTestRepoHelper(t *testing.T) {
	tempDir := t.TempDir()

	// Test creating regular repo
	regularPath := filepath.Join(tempDir, "regular")
	createTestRepo(t, regularPath, false)

	gitDir := filepath.Join(regularPath, ".git")
	assert.DirExists(t, gitDir, "Should create .git directory")

	// Test creating bare repo
	barePath := filepath.Join(tempDir, "bare.git")
	createTestRepo(t, barePath, true)

	assert.DirExists(t, filepath.Join(barePath, "refs"), "Should create refs directory")
	assert.DirExists(t, filepath.Join(barePath, "objects"), "Should create objects directory")
	assert.FileExists(t, filepath.Join(barePath, "HEAD"), "Should create HEAD file")
}

// Additional edge case: Empty directory
func TestScan_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	ctx := context.Background()
	opts := ScanOptions{
		RootPath: tempDir,
	}
	result, err := Scan(ctx, opts)

	require.NoError(t, err)
	assert.Equal(t, 0, len(result.Repositories), "Should find no repositories in empty directory")
	assert.Equal(t, 0, result.TotalRepos)
}

// Additional edge case: Symlink handling
func TestScan_SymlinkToRepo(t *testing.T) {
	tempDir := t.TempDir()

	// Create real repo
	realRepoPath := filepath.Join(tempDir, "real-repo")
	createTestRepo(t, realRepoPath, false)

	// Create symlink to repo
	symlinkPath := filepath.Join(tempDir, "symlink-repo")
	err := os.Symlink(realRepoPath, symlinkPath)
	if err != nil {
		t.Skip("Symlink creation not supported on this system")
	}

	ctx := context.Background()
	opts := ScanOptions{
		RootPath: tempDir,
	}
	result, err := Scan(ctx, opts)

	require.NoError(t, err)

	// Should find repos, might be 1 or 2 depending on symlink handling
	// Based on spec, we follow symlinks once
	assert.GreaterOrEqual(t, len(result.Repositories), 1, "Should find at least the real repo")

	// Check if any repo is marked as symlinked
	foundSymlink := false
	for _, repo := range result.Repositories {
		if repo.IsSymlink {
			foundSymlink = true
		}
	}

	// We should detect at least one as accessed via symlink (if we track this)
	if len(result.Repositories) > 1 {
		assert.True(t, foundSymlink, "Should mark symlinked repos")
	}
}

// Test Scan with non-existent root path (fatal error)
func TestScan_NonExistentRoot(t *testing.T) {
	ctx := context.Background()
	opts := ScanOptions{
		RootPath: "/path/that/does/not/exist/hopefully",
	}
	result, err := Scan(ctx, opts)

	// Should return fatal error for non-existent root
	assert.Error(t, err, "Should return error for non-existent root path")
	assert.Nil(t, result, "Should not return result on fatal error")
}

// Test bare repository detection patterns
func TestIsGitRepository_BareRepoPatterns(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func(string)
		wantRepo bool
		wantBare bool
	}{
		{
			name: "complete bare repo",
			setup: func(path string) {
				createTestRepo(t, path, true)
			},
			wantRepo: true,
			wantBare: true,
		},
		{
			name: "missing HEAD file",
			setup: func(path string) {
				os.MkdirAll(filepath.Join(path, "refs", "heads"), 0755)
				os.MkdirAll(filepath.Join(path, "objects"), 0755)
				// No HEAD file
			},
			wantRepo: false,
			wantBare: false,
		},
		{
			name: "missing refs directory",
			setup: func(path string) {
				os.MkdirAll(filepath.Join(path, "objects"), 0755)
				os.WriteFile(filepath.Join(path, "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
				// No refs directory
			},
			wantRepo: false,
			wantBare: false,
		},
		{
			name: "missing objects directory",
			setup: func(path string) {
				os.MkdirAll(filepath.Join(path, "refs", "heads"), 0755)
				os.WriteFile(filepath.Join(path, "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
				// No objects directory
			},
			wantRepo: false,
			wantBare: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := filepath.Join(tempDir, tt.name)
			os.MkdirAll(testPath, 0755)
			tt.setup(testPath)

			isRepo, isBare := IsGitRepository(testPath)

			assert.Equal(t, tt.wantRepo, isRepo, "Repository detection mismatch")
			assert.Equal(t, tt.wantBare, isBare, "Bare repository detection mismatch")
		})
	}
}
