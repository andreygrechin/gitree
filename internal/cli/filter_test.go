package cli

import (
	"testing"

	"github.com/andreygrechin/gitree/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestIsClean_AllConditionsMet verifies clean state when all 9 conditions are satisfied.
func TestIsClean_AllConditionsMet(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "main",
			IsDetached: false,
			HasRemote:  true,
			Ahead:      0,
			Behind:     0,
			HasStashes: false,
			HasChanges: false,
			Error:      "",
		},
	}

	assert.True(t, IsClean(repo), "Repository should be clean when all conditions are met")

	// Also test with "master" branch
	repo.GitStatus.Branch = "master"
	assert.True(t, IsClean(repo), "Repository on master branch should also be clean")
}

// TestIsClean_FeatureBranch verifies repo on feature branch is not clean.
func TestIsClean_FeatureBranch(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "feature-branch",
			IsDetached: false,
			HasRemote:  true,
			Ahead:      0,
			Behind:     0,
			HasStashes: false,
			HasChanges: false,
			Error:      "",
		},
	}

	assert.False(t, IsClean(repo), "Repository on feature branch is not clean")
}

// TestIsClean_UncommittedChanges verifies repo with uncommitted changes is not clean.
func TestIsClean_UncommittedChanges(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "main",
			IsDetached: false,
			HasRemote:  true,
			Ahead:      0,
			Behind:     0,
			HasStashes: false,
			HasChanges: true, // Uncommitted changes
			Error:      "",
		},
	}

	assert.False(t, IsClean(repo), "Repository with uncommitted changes is not clean")
}

// TestIsClean_HasStashes verifies repo with stashes is not clean.
func TestIsClean_HasStashes(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "main",
			IsDetached: false,
			HasRemote:  true,
			Ahead:      0,
			Behind:     0,
			HasStashes: true, // Has stashes
			HasChanges: false,
			Error:      "",
		},
	}

	assert.False(t, IsClean(repo), "Repository with stashes is not clean")
}

// TestIsClean_NoRemote verifies repo without remote is not clean.
func TestIsClean_NoRemote(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "main",
			IsDetached: false,
			HasRemote:  false, // No remote
			Ahead:      0,
			Behind:     0,
			HasStashes: false,
			HasChanges: false,
			Error:      "",
		},
	}

	assert.False(t, IsClean(repo), "Repository without remote is not clean")
}

// TestIsClean_AheadOfRemote verifies repo ahead of remote is not clean.
func TestIsClean_AheadOfRemote(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "main",
			IsDetached: false,
			HasRemote:  true,
			Ahead:      3, // Ahead of remote
			Behind:     0,
			HasStashes: false,
			HasChanges: false,
			Error:      "",
		},
	}

	assert.False(t, IsClean(repo), "Repository ahead of remote is not clean")
}

// TestIsClean_BehindRemote verifies repo behind remote is not clean.
func TestIsClean_BehindRemote(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "main",
			IsDetached: false,
			HasRemote:  true,
			Ahead:      0,
			Behind:     2, // Behind remote
			HasStashes: false,
			HasChanges: false,
			Error:      "",
		},
	}

	assert.False(t, IsClean(repo), "Repository behind remote is not clean")
}

// TestIsClean_DetachedHEAD verifies detached HEAD is not clean.
func TestIsClean_DetachedHEAD(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "DETACHED",
			IsDetached: true, // Detached HEAD
			HasRemote:  true,
			Ahead:      0,
			Behind:     0,
			HasStashes: false,
			HasChanges: false,
			Error:      "",
		},
	}

	assert.False(t, IsClean(repo), "Repository with detached HEAD is not clean")
}

// TestIsClean_StatusError verifies repo with status error is not clean (fail-safe).
func TestIsClean_StatusError(t *testing.T) {
	repo := &models.Repository{
		Path: "/test/repo",
		Name: "repo",
		GitStatus: &models.GitStatus{
			Branch:     "main",
			IsDetached: false,
			HasRemote:  true,
			Ahead:      0,
			Behind:     0,
			HasStashes: false,
			HasChanges: false,
			Error:      "timeout extracting status", // Error present
		},
	}

	assert.False(t, IsClean(repo), "Repository with status error is not clean (fail-safe)")
}

// TestIsClean_NilStatus verifies nil status is not clean (fail-safe).
func TestIsClean_NilStatus(t *testing.T) {
	repo := &models.Repository{
		Path:      "/test/repo",
		Name:      "repo",
		GitStatus: nil, // Nil status
	}

	assert.False(t, IsClean(repo), "Repository with nil status is not clean (fail-safe)")
}

// TestIsClean_NilRepository verifies nil repository is not clean (fail-safe).
func TestIsClean_NilRepository(t *testing.T) {
	assert.False(t, IsClean(nil), "Nil repository is not clean (fail-safe)")
}

// TestFilterRepositories_EmptyInput verifies empty input returns empty output.
func TestFilterRepositories_EmptyInput(t *testing.T) {
	repos := []*models.Repository{}

	// Test with ShowAll = false (default filtering)
	filtered := FilterRepositories(repos, FilterOptions{ShowAll: false})
	assert.Empty(t, filtered, "Empty input should return empty output")

	// Test with ShowAll = true
	all := FilterRepositories(repos, FilterOptions{ShowAll: true})
	assert.Empty(t, all, "Empty input should return empty output even with ShowAll")
}

// TestFilterRepositories_AllClean verifies all clean repos filtered out by default.
func TestFilterRepositories_AllClean(t *testing.T) {
	repos := []*models.Repository{
		{
			Path: "/test/repo1",
			Name: "repo1",
			GitStatus: &models.GitStatus{
				Branch: "main", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/repo2",
			Name: "repo2",
			GitStatus: &models.GitStatus{
				Branch: "master", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
	}

	// Default filtering (ShowAll = false) should show none
	filtered := FilterRepositories(repos, FilterOptions{ShowAll: false})
	assert.Empty(t, filtered, "All clean repos should be filtered out by default")

	// With ShowAll = true, should show all
	all := FilterRepositories(repos, FilterOptions{ShowAll: true})
	assert.Len(t, all, 2, "ShowAll should return all repos including clean ones")
}

// TestFilterRepositories_AllDirty verifies all dirty repos shown in both modes.
func TestFilterRepositories_AllDirty(t *testing.T) {
	repos := []*models.Repository{
		{
			Path: "/test/repo1",
			Name: "repo1",
			GitStatus: &models.GitStatus{
				Branch: "feature", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/repo2",
			Name: "repo2",
			GitStatus: &models.GitStatus{
				Branch: "main", IsDetached: false, HasRemote: true,
				Ahead: 3, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
	}

	// Default filtering should show all (all need attention)
	filtered := FilterRepositories(repos, FilterOptions{ShowAll: false})
	assert.Len(t, filtered, 2, "All dirty repos should be shown by default")

	// ShowAll should also show all
	all := FilterRepositories(repos, FilterOptions{ShowAll: true})
	assert.Len(t, all, 2, "ShowAll should return all repos")
}

// TestFilterRepositories_Mixed verifies mixed clean/dirty repos filtered correctly.
func TestFilterRepositories_Mixed(t *testing.T) {
	repos := []*models.Repository{
		{
			Path: "/test/clean1",
			Name: "clean1",
			GitStatus: &models.GitStatus{
				Branch: "main", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/dirty1",
			Name: "dirty1",
			GitStatus: &models.GitStatus{
				Branch: "feature", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/clean2",
			Name: "clean2",
			GitStatus: &models.GitStatus{
				Branch: "master", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/dirty2",
			Name: "dirty2",
			GitStatus: &models.GitStatus{
				Branch: "main", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: true, Error: "",
			},
		},
	}

	// Default filtering should show only dirty repos
	filtered := FilterRepositories(repos, FilterOptions{ShowAll: false})
	assert.Len(t, filtered, 2, "Should filter to show only dirty repos")
	assert.Equal(t, "dirty1", filtered[0].Name, "First dirty repo should be dirty1")
	assert.Equal(t, "dirty2", filtered[1].Name, "Second dirty repo should be dirty2")

	// ShowAll should show all 4 repos
	all := FilterRepositories(repos, FilterOptions{ShowAll: true})
	assert.Len(t, all, 4, "ShowAll should return all 4 repos")
}

// TestFilterRepositories_OrderPreservation verifies order is maintained.
func TestFilterRepositories_OrderPreservation(t *testing.T) {
	repos := []*models.Repository{
		{
			Path: "/test/a",
			Name: "a",
			GitStatus: &models.GitStatus{
				Branch: "feature-a", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/b",
			Name: "b",
			GitStatus: &models.GitStatus{
				Branch: "feature-b", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/c",
			Name: "c",
			GitStatus: &models.GitStatus{
				Branch: "feature-c", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
	}

	filtered := FilterRepositories(repos, FilterOptions{ShowAll: false})
	assert.Len(t, filtered, 3, "Should return all 3 dirty repos")
	assert.Equal(t, "a", filtered[0].Name, "Order should be preserved: a first")
	assert.Equal(t, "b", filtered[1].Name, "Order should be preserved: b second")
	assert.Equal(t, "c", filtered[2].Name, "Order should be preserved: c third")
}

// TestFilterRepositories_NonModification verifies original slice not modified.
func TestFilterRepositories_NonModification(t *testing.T) {
	repos := []*models.Repository{
		{
			Path: "/test/clean",
			Name: "clean",
			GitStatus: &models.GitStatus{
				Branch: "main", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
		{
			Path: "/test/dirty",
			Name: "dirty",
			GitStatus: &models.GitStatus{
				Branch: "feature", IsDetached: false, HasRemote: true,
				Ahead: 0, Behind: 0, HasStashes: false, HasChanges: false, Error: "",
			},
		},
	}

	// Filter repos
	filtered := FilterRepositories(repos, FilterOptions{ShowAll: false})

	// Verify original slice not modified
	assert.Len(t, repos, 2, "Original slice should still have 2 repos")
	assert.Equal(t, "clean", repos[0].Name, "Original slice first element unchanged")
	assert.Equal(t, "dirty", repos[1].Name, "Original slice second element unchanged")

	// Verify filtered slice is different
	assert.Len(t, filtered, 1, "Filtered slice should have 1 repo")
	assert.Equal(t, "dirty", filtered[0].Name, "Filtered slice should contain dirty repo")
}
