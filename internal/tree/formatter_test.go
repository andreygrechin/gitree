package tree

import (
	"errors"
	"strings"
	"testing"

	"github.com/andreygrechin/gitree/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T054: Test Build() creating correct tree structure from flat repo list.
func TestBuild_CreatesCorrectTreeStructure(t *testing.T) {
	repos := []*models.Repository{
		{Path: "/root/project1", Name: "project1"},
		{Path: "/root/project2", Name: "project2"},
		{Path: "/root/nested/project3", Name: "project3"},
	}

	root := Build("/root", repos, nil)

	require.NotNil(t, root)
	assert.Len(t, root.Children, 3) // project1, project2, and nested directory
}

// T055: Test Build() calculating relative paths correctly.
func TestBuild_CalculatesRelativePaths(t *testing.T) {
	repos := []*models.Repository{
		{Path: "/root/project1", Name: "project1"},
		{Path: "/root/a/b/project2", Name: "project2"},
	}

	root := Build("/root", repos, nil)

	require.NotNil(t, root)
	// Check relative paths are calculated correctly
	var relPaths []string
	collectRelativePaths(root, &relPaths)

	assert.Contains(t, relPaths, "project1")
	assert.Contains(t, relPaths, "a/b/project2")
}

// Helper function to collect relative paths from tree.
func collectRelativePaths(node *models.TreeNode, paths *[]string) {
	if node.Repository != nil {
		*paths = append(*paths, node.RelativePath)
	}
	for _, child := range node.Children {
		collectRelativePaths(child, paths)
	}
}

// T056: Test Build() sorting children alphabetically.
func TestBuild_SortsChildrenAlphabetically(t *testing.T) {
	repos := []*models.Repository{
		{Path: "/root/zebra", Name: "zebra"},
		{Path: "/root/apple", Name: "apple"},
		{Path: "/root/middle", Name: "middle"},
	}

	root := Build("/root", repos, nil)

	require.NotNil(t, root)
	require.Len(t, root.Children, 3)

	// Check alphabetical order
	assert.Equal(t, "apple", root.Children[0].Repository.Name)
	assert.Equal(t, "middle", root.Children[1].Repository.Name)
	assert.Equal(t, "zebra", root.Children[2].Repository.Name)
}

// T057: Test Build() setting depth levels correctly.
func TestBuild_SetsDepthLevels(t *testing.T) {
	repos := []*models.Repository{
		{Path: "/root/project1", Name: "project1"},
		{Path: "/root/a/project2", Name: "project2"},
		{Path: "/root/a/b/project3", Name: "project3"},
	}

	root := Build("/root", repos, nil)

	require.NotNil(t, root)

	// Root should be depth 0
	assert.Equal(t, 0, root.Depth)

	// Find and check depths
	var depths []int
	collectDepths(root, &depths)

	// Should have depths: 0 (root), 1 (project1, a), 2 (project2, b), 3 (project3)
	assert.Contains(t, depths, 0)
	assert.Contains(t, depths, 1)
	assert.Contains(t, depths, 2)
}

// Helper to collect all depths.
func collectDepths(node *models.TreeNode, depths *[]int) {
	*depths = append(*depths, node.Depth)
	for _, child := range node.Children {
		collectDepths(child, depths)
	}
}

// T058: Test Build() marking IsLast flags for last children.
func TestBuild_MarksIsLastFlags(t *testing.T) {
	repos := []*models.Repository{
		{Path: "/root/project1", Name: "project1"},
		{Path: "/root/project2", Name: "project2"},
		{Path: "/root/project3", Name: "project3"},
	}

	root := Build("/root", repos, nil)

	require.NotNil(t, root)
	require.Len(t, root.Children, 3)

	// First two should NOT be last
	assert.False(t, root.Children[0].IsLast)
	assert.False(t, root.Children[1].IsLast)

	// Last one should be marked as last
	assert.True(t, root.Children[2].IsLast)
}

// T059: Test Format() with single repository.
func TestFormat_SingleRepository(t *testing.T) {
	repos := []*models.Repository{
		{
			Path: "/root/project",
			Name: "project",
			GitStatus: &models.GitStatus{
				Branch:    "main",
				HasRemote: true,
			},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	assert.Contains(t, output, "project")
	assert.Contains(t, output, "[[ main")
	assert.Contains(t, output, "]]")
}

// T060: Test Format() with multiple repos at same level.
func TestFormat_MultipleReposSameLevel(t *testing.T) {
	repos := []*models.Repository{
		{
			Path:      "/root/project1",
			Name:      "project1",
			GitStatus: &models.GitStatus{Branch: "main", HasRemote: true},
		},
		{
			Path:      "/root/project2",
			Name:      "project2",
			GitStatus: &models.GitStatus{Branch: "develop", HasRemote: true},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	assert.Contains(t, output, "project1")
	assert.Contains(t, output, "project2")
	assert.Contains(t, output, "[[ main")
	assert.Contains(t, output, "[[ develop")
}

// T061: Test Format() with nested repos.
func TestFormat_NestedRepos(t *testing.T) {
	repos := []*models.Repository{
		{
			Path:      "/root/project1",
			Name:      "project1",
			GitStatus: &models.GitStatus{Branch: "main"},
		},
		{
			Path:      "/root/nested/project2",
			Name:      "project2",
			GitStatus: &models.GitStatus{Branch: "develop"},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	assert.Contains(t, output, "project1")
	assert.Contains(t, output, "nested")
	assert.Contains(t, output, "project2")
}

// T062: Test Format() using correct connectors.
func TestFormat_UsesCorrectConnectors(t *testing.T) {
	repos := []*models.Repository{
		{
			Path:      "/root/project1",
			Name:      "project1",
			GitStatus: &models.GitStatus{Branch: "main"},
		},
		{
			Path:      "/root/project2",
			Name:      "project2",
			GitStatus: &models.GitStatus{Branch: "develop"},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	// Should use tree connectors
	assert.Contains(t, output, "├──") // Non-last child connector
	assert.Contains(t, output, "└──") // Last child connector
}

// T063: Test Format() including Git status inline.
func TestFormat_IncludesGitStatusInline(t *testing.T) {
	repos := []*models.Repository{
		{
			Path: "/root/project",
			Name: "project",
			GitStatus: &models.GitStatus{
				Branch:     "main",
				Ahead:      2,
				Behind:     1,
				HasRemote:  true,
				HasStashes: true,
				HasChanges: true,
			},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	// Check that status is formatted inline
	assert.Contains(t, output, "project")
	assert.Contains(t, output, "[[ main")
	assert.Contains(t, output, "↑2")
	assert.Contains(t, output, "↓1")
	assert.Contains(t, output, "$")
	assert.Contains(t, output, "*")
}

// T064: Test Format() matching examples from cli-contract.md.
func TestFormat_MatchesCliContractExamples(t *testing.T) {
	// Example: Simple tree with status
	repos := []*models.Repository{
		{
			Path: "/root/proj1",
			Name: "proj1",
			GitStatus: &models.GitStatus{
				Branch:    "main",
				HasRemote: false,
			},
		},
		{
			Path: "/root/proj2",
			Name: "proj2",
			GitStatus: &models.GitStatus{
				Branch:    "develop",
				HasRemote: true,
				Ahead:     2,
			},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	// Should match general structure
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Greater(t, len(lines), 1)

	// Check for proper formatting
	assert.Contains(t, output, "proj1")
	assert.Contains(t, output, "proj2")
	assert.Contains(t, output, "○")  // No remote indicator for proj1
	assert.Contains(t, output, "↑2") // Ahead indicator for proj2
}

// T065: Test Format() with empty repository list.
func TestFormat_EmptyRepositoryList(t *testing.T) {
	repos := []*models.Repository{}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	// Should handle empty list gracefully
	assert.NotEmpty(t, output)
	// Typically would show root or "No repositories found" message
}

var ErrCorruptedRepository = errors.New("corrupted repository")

// Additional test: Format with error indicator.
func TestFormat_WithErrorIndicator(t *testing.T) {
	repos := []*models.Repository{
		{
			Path:  "/root/broken",
			Name:  "broken",
			Error: ErrCorruptedRepository,
			GitStatus: &models.GitStatus{
				Branch: "main",
				Error:  "corrupted",
			},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	assert.Contains(t, output, "broken")
	assert.Contains(t, output, "error")
}

// Additional test: Format with timeout indicator.
func TestFormat_WithTimeoutIndicator(t *testing.T) {
	repos := []*models.Repository{
		{
			Path:       "/root/slow",
			Name:       "slow",
			HasTimeout: true,
			GitStatus: &models.GitStatus{
				Branch: "main",
				Error:  "timeout",
			},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	assert.Contains(t, output, "slow")
	assert.Contains(t, output, "timeout")
}

// Additional test: Format with bare repository.
func TestFormat_WithBareRepository(t *testing.T) {
	repos := []*models.Repository{
		{
			Path:   "/root/bare.git",
			Name:   "bare.git",
			IsBare: true,
			GitStatus: &models.GitStatus{
				Branch: "main",
			},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	assert.Contains(t, output, "bare.git")
	assert.Contains(t, output, "bare")
}

// Test Build with nil repository list.
func TestBuild_NilRepositoryList(t *testing.T) {
	root := Build("/root", nil, nil)
	assert.NotNil(t, root)
}

// Test Format with vertical pipe for non-last children.
func TestFormat_VerticalPipeForNonLastChildren(t *testing.T) {
	repos := []*models.Repository{
		{
			Path:      "/root/project1",
			Name:      "project1",
			GitStatus: &models.GitStatus{Branch: "main", HasRemote: true},
		},
		{
			Path:      "/root/dir/project2",
			Name:      "project2",
			GitStatus: &models.GitStatus{Branch: "develop", HasRemote: true},
		},
		{
			Path:      "/root/dir/project3",
			Name:      "project3",
			GitStatus: &models.GitStatus{Branch: "feature", HasRemote: true},
		},
	}

	root := Build("/root", repos, nil)
	output := Format(root, nil)

	// Should contain vertical pipe for continuation
	// When there's dir as non-last child with nested projects
	assert.Contains(t, output, "│")
}
