package models

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ErrFakeError = errors.New("fake error for testing")

// T008: Test Repository struct validation.
func TestRepositoryValidation(t *testing.T) {
	tests := []struct {
		name        string
		repo        Repository
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid repository",
			repo: Repository{
				Path: "/home/user/project",
				Name: "project",
			},
			expectError: false,
		},
		{
			name: "empty path",
			repo: Repository{
				Name: "project",
			},
			expectError: true,
			errorMsg:    "path cannot be empty",
		},
		{
			name: "empty name",
			repo: Repository{
				Path: "/home/user/project",
			},
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name: "relative path",
			repo: Repository{
				Path: "relative/path",
				Name: "project",
			},
			expectError: true,
			errorMsg:    "path must be absolute",
		},
		{
			name: "bare repo with changes - invalid",
			repo: Repository{
				Path:   "/home/user/project",
				Name:   "project",
				IsBare: true,
				GitStatus: &GitStatus{
					Branch:     "main",
					HasChanges: true,
				},
			},
			expectError: true,
			errorMsg:    "bare repository cannot have uncommitted changes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// T009: Test GitStatus struct validation.
func TestGitStatusValidation(t *testing.T) {
	tests := []struct {
		name        string
		status      GitStatus
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid status",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     2,
				Behind:    1,
			},
			expectError: false,
		},
		{
			name: "empty branch",
			status: GitStatus{
				Branch: "",
			},
			expectError: true,
			errorMsg:    "branch cannot be empty",
		},
		{
			name: "detached HEAD with wrong branch name",
			status: GitStatus{
				Branch:     "main",
				IsDetached: true,
			},
			expectError: true,
			errorMsg:    "detached HEAD must have branch = 'DETACHED'",
		},
		{
			name: "detached HEAD correct",
			status: GitStatus{
				Branch:     "DETACHED",
				IsDetached: true,
			},
			expectError: false,
		},
		{
			name: "no remote but has ahead/behind",
			status: GitStatus{
				Branch:    "main",
				HasRemote: false,
				Ahead:     2,
			},
			expectError: true,
			errorMsg:    "no remote but ahead/behind counts are non-zero",
		},
		{
			name: "negative ahead count",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     -1,
			},
			expectError: true,
			errorMsg:    "ahead/behind counts cannot be negative",
		},
		{
			name: "negative behind count",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Behind:    -2,
			},
			expectError: true,
			errorMsg:    "ahead/behind counts cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// T010: Test TreeNode struct validation.
func TestTreeNodeValidation(t *testing.T) {
	tests := []struct {
		name        string
		node        TreeNode
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid tree node",
			node: TreeNode{
				Repository: &Repository{
					Path: "/home/user/project",
					Name: "project",
				},
				Depth:        0,
				RelativePath: "project",
			},
			expectError: false,
		},
		{
			name: "nil repository",
			node: TreeNode{
				Depth:        0,
				RelativePath: "project",
			},
			expectError: true,
			errorMsg:    "repository cannot be nil",
		},
		{
			name: "negative depth",
			node: TreeNode{
				Repository: &Repository{
					Path: "/home/user/project",
					Name: "project",
				},
				Depth:        -1,
				RelativePath: "project",
			},
			expectError: true,
			errorMsg:    "depth cannot be negative",
		},
		{
			name: "empty relative path",
			node: TreeNode{
				Repository: &Repository{
					Path: "/home/user/project",
					Name: "project",
				},
				Depth: 0,
			},
			expectError: true,
			errorMsg:    "relative path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// T010: Test TreeNode methods.
func TestTreeNodeMethods(t *testing.T) {
	t.Run("AddChild sets depth correctly", func(t *testing.T) {
		parent := &TreeNode{
			Repository: &Repository{
				Path: "/home/user",
				Name: "user",
			},
			Depth:        0,
			RelativePath: ".",
		}

		child := &TreeNode{
			Repository: &Repository{
				Path: "/home/user/project",
				Name: "project",
			},
			RelativePath: "project",
		}

		parent.AddChild(child)

		assert.Equal(t, 1, child.Depth)
		assert.Len(t, parent.Children, 1)
		assert.Equal(t, child, parent.Children[0])
	})

	t.Run("SortChildren sorts alphabetically and sets IsLast", func(t *testing.T) {
		parent := &TreeNode{
			Repository: &Repository{
				Path: "/home/user",
				Name: "user",
			},
			Depth:        0,
			RelativePath: ".",
		}

		child1 := &TreeNode{
			Repository: &Repository{
				Path: "/home/user/zulu",
				Name: "zulu",
			},
			RelativePath: "zulu",
		}

		child2 := &TreeNode{
			Repository: &Repository{
				Path: "/home/user/alpha",
				Name: "alpha",
			},
			RelativePath: "alpha",
		}

		child3 := &TreeNode{
			Repository: &Repository{
				Path: "/home/user/bravo",
				Name: "bravo",
			},
			RelativePath: "bravo",
		}

		parent.Children = []*TreeNode{child1, child2, child3}
		parent.SortChildren()

		require.Len(t, parent.Children, 3)
		assert.Equal(t, "alpha", parent.Children[0].Repository.Name)
		assert.Equal(t, "bravo", parent.Children[1].Repository.Name)
		assert.Equal(t, "zulu", parent.Children[2].Repository.Name)

		assert.False(t, parent.Children[0].IsLast)
		assert.False(t, parent.Children[1].IsLast)
		assert.True(t, parent.Children[2].IsLast)
	})
}

// T011: Test ScanResult struct validation.
func TestScanResultValidation(t *testing.T) {
	tests := []struct {
		name        string
		result      ScanResult
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid scan result",
			result: ScanResult{
				RootPath:     "/home/user",
				Repositories: []*Repository{},
				TotalScanned: 10,
				TotalRepos:   0,
				Duration:     100 * time.Millisecond,
			},
			expectError: false,
		},
		{
			name: "empty root path",
			result: ScanResult{
				Repositories: []*Repository{},
			},
			expectError: true,
			errorMsg:    "root path cannot be empty",
		},
		{
			name: "relative root path",
			result: ScanResult{
				RootPath:     "relative/path",
				Repositories: []*Repository{},
			},
			expectError: true,
			errorMsg:    "root path must be absolute",
		},
		{
			name: "nil repositories slice",
			result: ScanResult{
				RootPath: "/home/user",
			},
			expectError: true,
			errorMsg:    "repositories slice cannot be nil",
		},
		{
			name: "total repos mismatch",
			result: ScanResult{
				RootPath: "/home/user",
				Repositories: []*Repository{
					{Path: "/home/user/project", Name: "project"},
				},
				TotalRepos: 2, // Mismatch
			},
			expectError: true,
			errorMsg:    "total repos mismatch",
		},
		{
			name: "total scanned less than total repos",
			result: ScanResult{
				RootPath: "/home/user",
				Repositories: []*Repository{
					{Path: "/home/user/project", Name: "project"},
				},
				TotalScanned: 0,
				TotalRepos:   1,
			},
			expectError: true,
			errorMsg:    "total scanned < total repos",
		},
		{
			name: "negative duration",
			result: ScanResult{
				RootPath:     "/home/user",
				Repositories: []*Repository{},
				Duration:     -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "duration cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// T011: Test ScanResult helper methods.
func TestScanResultMethods(t *testing.T) {
	t.Run("HasErrors returns true when errors exist", func(t *testing.T) {
		result := ScanResult{
			Errors: []error{fmt.Errorf("test error: %w", ErrFakeError)},
		}
		assert.True(t, result.HasErrors())
	})

	t.Run("HasErrors returns false when no errors", func(t *testing.T) {
		result := ScanResult{
			Errors: []error{},
		}
		assert.False(t, result.HasErrors())
	})

	t.Run("SuccessRate with no repos", func(t *testing.T) {
		result := ScanResult{
			TotalRepos:   0,
			Repositories: []*Repository{},
		}
		assert.InDelta(t, 1.0, result.SuccessRate(), 1e-9)
	})

	t.Run("SuccessRate with all successful repos", func(t *testing.T) {
		result := ScanResult{
			TotalRepos: 3,
			Repositories: []*Repository{
				{Path: "/home/user/p1", Name: "p1"},
				{Path: "/home/user/p2", Name: "p2"},
				{Path: "/home/user/p3", Name: "p3"},
			},
		}
		assert.InDelta(t, 1.0, result.SuccessRate(), 1e-9)
	})

	t.Run("SuccessRate with some failed repos", func(t *testing.T) {
		result := ScanResult{
			TotalRepos: 4,
			Repositories: []*Repository{
				{Path: "/home/user/p1", Name: "p1"},
				{Path: "/home/user/p2", Name: "p2", Error: fmt.Errorf("error: %w", ErrFakeError)},
				{Path: "/home/user/p3", Name: "p3", HasTimeout: true},
				{Path: "/home/user/p4", Name: "p4"},
			},
		}
		assert.InDelta(t, 0.5, result.SuccessRate(), 1e-9)
	})
}

// T012: Test GitStatus.Format() method (now with colorization and double brackets).
func TestGitStatusFormat(t *testing.T) {
	// Disable colors for these baseline tests to check structure
	color.NoColor = true
	defer func() { color.NoColor = false }()

	tests := []struct {
		name     string
		status   GitStatus
		expected string
	}{
		{
			name: "simple branch in sync",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
			},
			expected: "[[ main ]]",
		},
		{
			name: "ahead of remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     2,
			},
			expected: "[[ main | ↑2 ]]",
		},
		{
			name: "behind remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Behind:    1,
			},
			expected: "[[ main | ↓1 ]]",
		},
		{
			name: "ahead and behind",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     2,
				Behind:    1,
			},
			expected: "[[ main | ↑2 ↓1 ]]",
		},
		{
			name: "with stashes",
			status: GitStatus{
				Branch:     "develop",
				HasRemote:  true,
				HasStashes: true,
			},
			expected: "[[ develop | $ ]]",
		},
		{
			name: "with uncommitted changes",
			status: GitStatus{
				Branch:     "main",
				HasRemote:  true,
				HasChanges: true,
			},
			expected: "[[ main | * ]]",
		},
		{
			name: "with stashes and changes",
			status: GitStatus{
				Branch:     "develop",
				HasRemote:  true,
				HasStashes: true,
				HasChanges: true,
			},
			expected: "[[ develop | $ * ]]",
		},
		{
			name: "detached HEAD",
			status: GitStatus{
				Branch:     "DETACHED",
				IsDetached: true,
				HasRemote:  false,
			},
			expected: "[[ DETACHED | ○ ]]",
		},
		{
			name: "no remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: false,
			},
			expected: "[[ main | ○ ]]",
		},
		{
			name: "all indicators",
			status: GitStatus{
				Branch:     "feature",
				HasRemote:  true,
				Ahead:      3,
				Behind:     2,
				HasStashes: true,
				HasChanges: true,
			},
			expected: "[[ feature | ↑3 ↓2 $ * ]]",
		},
		{
			name: "ahead exceeded limit",
			status: GitStatus{
				Branch:    "feature",
				HasRemote: true,
				Ahead:     maxCommitsToCount + 1,
			},
			expected: "[[ feature | ↑99+ ]]",
		},
		{
			name: "behind exceeded limit",
			status: GitStatus{
				Branch:    "feature",
				HasRemote: true,
				Behind:    maxCommitsToCount + 1,
			},
			expected: "[[ feature | ↓99+ ]]",
		},
		{
			name: "both ahead and behind exceeded",
			status: GitStatus{
				Branch:    "feature",
				HasRemote: true,
				Ahead:     maxCommitsToCount + 1,
				Behind:    maxCommitsToCount + 1,
			},
			expected: "[[ feature | ↑99+ ↓99+ ]]",
		},
		{
			name: "with error",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Error:     "partial failure",
			},
			expected: "[[ main | error ]]",
		},
		{
			name: "with error and partial info",
			status: GitStatus{
				Branch:     "main",
				HasRemote:  true,
				Ahead:      2,
				HasChanges: true,
				Error:      "timeout",
			},
			expected: "[[ main | ↑2 * error ]]",
		},
		{
			name: "with N/A branch and error",
			status: GitStatus{
				Branch: "N/A",
				Error:  "failed to extract branch",
			},
			expected: "[[ N/A | error ]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.Format()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// === User Story 1: Distinguish Repository Metadata from Names ===

// T009 [US1]: Verify output uses double brackets [[ ]] instead of [ ].
func TestGitStatusFormatDoubleBrackets(t *testing.T) {
	color.NoColor = false // Enable colors
	status := GitStatus{
		Branch:    "main",
		HasRemote: true,
	}
	output := status.Format()

	assert.Contains(t, output, "[[", "Output should contain opening double bracket")
	assert.Contains(t, output, "]]", "Output should contain closing double bracket")
	assert.NotContains(t, output, "[main]", "Output should not contain old single bracket format")
}

// T010 [US1]: Verify bracket characters contain ANSI code \033[90;1m (bold gray) when colors enabled.
func TestGitStatusFormatGrayBrackets(t *testing.T) {
	color.NoColor = false // Enable colors
	status := GitStatus{
		Branch:    "main",
		HasRemote: true,
	}
	output := status.Format()

	// Bold gray color code is \033[90;1m (Hi-intensity black/gray with bold)
	assert.Contains(t, output, "\033[90;1m", "Output should contain ANSI bold gray color code")
	// Double brackets should be present
	assert.Contains(t, output, "[[")
	assert.Contains(t, output, "]]")
}

// T011 [US1]: Verify [[ ]] present but no ANSI codes when color.NoColor = true.
func TestGitStatusFormatDoubleBracketsNoColor(t *testing.T) {
	color.NoColor = true                     // Disable colors
	defer func() { color.NoColor = false }() // Restore after test

	status := GitStatus{
		Branch:    "main",
		HasRemote: true,
	}
	output := status.Format()

	// Double brackets should still be present
	assert.Contains(t, output, "[[", "Output should contain double brackets even without color")
	assert.Contains(t, output, "]]", "Output should contain double brackets even without color")
	// No ANSI codes should be present
	assert.NotContains(t, output, "\033[", "Output should not contain ANSI color codes when NoColor is true")
}

// Test IsStandardStatus method.
func TestIsStandardStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   GitStatus
		expected bool
	}{
		{
			name: "standard status - main with remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
			},
			expected: true,
		},
		{
			name: "standard status - master with remote",
			status: GitStatus{
				Branch:    "master",
				HasRemote: true,
			},
			expected: true,
		},
		{
			name: "non-standard - feature branch",
			status: GitStatus{
				Branch:    "feature/new-feature",
				HasRemote: true,
			},
			expected: false,
		},
		{
			name: "non-standard - main with no remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: false,
			},
			expected: false,
		},
		{
			name: "non-standard - main ahead",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     2,
			},
			expected: false,
		},
		{
			name: "non-standard - main behind",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Behind:    1,
			},
			expected: false,
		},
		{
			name: "non-standard - main with stashes",
			status: GitStatus{
				Branch:     "main",
				HasRemote:  true,
				HasStashes: true,
			},
			expected: false,
		},
		{
			name: "non-standard - main with uncommitted changes",
			status: GitStatus{
				Branch:     "main",
				HasRemote:  true,
				HasChanges: true,
			},
			expected: false,
		},
		{
			name: "non-standard - main with error",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Error:     "some error",
			},
			expected: false,
		},
		{
			name: "non-standard - detached HEAD",
			status: GitStatus{
				Branch:     "DETACHED",
				IsDetached: true,
				HasRemote:  false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsStandardStatus()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test yellow brackets for non-standard status.
func TestGitStatusFormatYellowBracketsNonStandard(t *testing.T) {
	color.NoColor = false // Enable colors

	tests := []struct {
		name   string
		status GitStatus
	}{
		{
			name: "feature branch",
			status: GitStatus{
				Branch:    "feature/new",
				HasRemote: true,
			},
		},
		{
			name: "main with no remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: false,
			},
		},
		{
			name: "main ahead",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     2,
			},
		},
		{
			name: "main behind",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Behind:    1,
			},
		},
		{
			name: "main with uncommitted changes",
			status: GitStatus{
				Branch:     "main",
				HasRemote:  true,
				HasChanges: true,
			},
		},
		{
			name: "main with error",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Error:     "some error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.status.Format()
			// Bold yellow color code is \033[33;1m
			assert.Contains(t, output, "\033[33;1m", "Output should contain ANSI bold yellow color code for non-standard status")
		})
	}
}

// Test gray brackets for standard status.
func TestGitStatusFormatGrayBracketsStandard(t *testing.T) {
	color.NoColor = false // Enable colors

	tests := []struct {
		name   string
		status GitStatus
	}{
		{
			name: "main with remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
			},
		},
		{
			name: "master with remote",
			status: GitStatus{
				Branch:    "master",
				HasRemote: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.status.Format()
			// Bold gray color code is \033[90;1m (Hi-intensity black/gray with bold)
			assert.Contains(t, output, "\033[90;1m", "Output should contain ANSI bold gray color code for standard status")
			// Should NOT contain yellow
			assert.NotContains(t, output, "\033[33;1m", "Output should not contain yellow color code for standard status")
		})
	}
}
