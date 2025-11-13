package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T008: Test Repository struct validation
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

// T009: Test GitStatus struct validation
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

// T010: Test TreeNode struct validation
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

// T010: Test TreeNode methods
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

// T011: Test ScanResult struct validation
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
				Tree: &TreeNode{
					Repository: &Repository{
						Path: "/home/user",
						Name: "user",
					},
					RelativePath: ".",
				},
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
				Tree: &TreeNode{
					Repository: &Repository{
						Path: "/home/user",
						Name: "user",
					},
					RelativePath: ".",
				},
			},
			expectError: true,
			errorMsg:    "root path cannot be empty",
		},
		{
			name: "relative root path",
			result: ScanResult{
				RootPath:     "relative/path",
				Repositories: []*Repository{},
				Tree: &TreeNode{
					Repository: &Repository{
						Path: "/home/user",
						Name: "user",
					},
					RelativePath: ".",
				},
			},
			expectError: true,
			errorMsg:    "root path must be absolute",
		},
		{
			name: "nil repositories slice",
			result: ScanResult{
				RootPath: "/home/user",
				Tree: &TreeNode{
					Repository: &Repository{
						Path: "/home/user",
						Name: "user",
					},
					RelativePath: ".",
				},
			},
			expectError: true,
			errorMsg:    "repositories slice cannot be nil",
		},
		{
			name: "nil tree",
			result: ScanResult{
				RootPath:     "/home/user",
				Repositories: []*Repository{},
			},
			expectError: true,
			errorMsg:    "tree cannot be nil",
		},
		{
			name: "total repos mismatch",
			result: ScanResult{
				RootPath: "/home/user",
				Repositories: []*Repository{
					{Path: "/home/user/project", Name: "project"},
				},
				Tree: &TreeNode{
					Repository: &Repository{
						Path: "/home/user",
						Name: "user",
					},
					RelativePath: ".",
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
				Tree: &TreeNode{
					Repository: &Repository{
						Path: "/home/user",
						Name: "user",
					},
					RelativePath: ".",
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
				Tree: &TreeNode{
					Repository: &Repository{
						Path: "/home/user",
						Name: "user",
					},
					RelativePath: ".",
				},
				Duration: -1 * time.Second,
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

// T011: Test ScanResult helper methods
func TestScanResultMethods(t *testing.T) {
	t.Run("HasErrors returns true when errors exist", func(t *testing.T) {
		result := ScanResult{
			Errors: []error{fmt.Errorf("test error")},
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
		assert.Equal(t, 1.0, result.SuccessRate())
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
		assert.Equal(t, 1.0, result.SuccessRate())
	})

	t.Run("SuccessRate with some failed repos", func(t *testing.T) {
		result := ScanResult{
			TotalRepos: 4,
			Repositories: []*Repository{
				{Path: "/home/user/p1", Name: "p1"},
				{Path: "/home/user/p2", Name: "p2", Error: fmt.Errorf("error")},
				{Path: "/home/user/p3", Name: "p3", HasTimeout: true},
				{Path: "/home/user/p4", Name: "p4"},
			},
		}
		assert.Equal(t, 0.5, result.SuccessRate())
	})
}

// T012: Test GitStatus.Format() method
func TestGitStatusFormat(t *testing.T) {
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
			expected: "[main]",
		},
		{
			name: "ahead of remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     2,
			},
			expected: "[main ↑2]",
		},
		{
			name: "behind remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Behind:    1,
			},
			expected: "[main ↓1]",
		},
		{
			name: "ahead and behind",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Ahead:     2,
				Behind:    1,
			},
			expected: "[main ↑2 ↓1]",
		},
		{
			name: "with stashes",
			status: GitStatus{
				Branch:     "develop",
				HasRemote:  true,
				HasStashes: true,
			},
			expected: "[develop $]",
		},
		{
			name: "with uncommitted changes",
			status: GitStatus{
				Branch:     "main",
				HasRemote:  true,
				HasChanges: true,
			},
			expected: "[main *]",
		},
		{
			name: "with stashes and changes",
			status: GitStatus{
				Branch:     "develop",
				HasRemote:  true,
				HasStashes: true,
				HasChanges: true,
			},
			expected: "[develop $ *]",
		},
		{
			name: "detached HEAD",
			status: GitStatus{
				Branch:     "DETACHED",
				IsDetached: true,
				HasRemote:  false,
			},
			expected: "[DETACHED ○]",
		},
		{
			name: "no remote",
			status: GitStatus{
				Branch:    "main",
				HasRemote: false,
			},
			expected: "[main ○]",
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
			expected: "[feature ↑3 ↓2 $ *]",
		},
		{
			name: "with error",
			status: GitStatus{
				Branch:    "main",
				HasRemote: true,
				Error:     "partial failure",
			},
			expected: "[main] error",
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
			expected: "[main ↑2 *] error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.Format()
			assert.Equal(t, tt.expected, result)
		})
	}
}
