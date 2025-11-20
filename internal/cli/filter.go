package cli

import "github.com/andreygrechin/gitree/internal/models"

// FilterOptions configures repository filtering behavior.
type FilterOptions struct {
	ShowAll bool // When true, disables filtering (shows all repos including clean ones). Default: false.
}

// IsClean determines if a repository is in a clean state per FR-008.
// A repository is considered clean if ALL of the following conditions are met:
// 1. On main or master branch
// 2. No uncommitted changes
// 3. No stashes
// 4. Has remote tracking configured
// 5. Not ahead of remote
// 6. Not behind remote
// 7. Not in detached HEAD state
// 8. No error in status extraction
//
// If any condition fails, the repository needs attention and is NOT clean.
//
// This function delegates to GitStatus.IsStandardStatus() to avoid code duplication.
func IsClean(repo *models.Repository) bool {
	// Fail-safe: nil status = unknown = not clean (FR-009)
	if repo == nil || repo.GitStatus == nil {
		return false
	}

	// Delegate to IsStandardStatus which checks all the clean state conditions
	return repo.GitStatus.IsStandardStatus()
}

// FilterRepositories filters the repository list based on options.
// By default (ShowAll=false), returns only repositories needing attention (not clean).
// With ShowAll=true, returns all repositories unchanged.
//
// The function preserves the original order of repositories and does not
// modify the input slice.
func FilterRepositories(repos []*models.Repository, opts FilterOptions) []*models.Repository {
	// If ShowAll is true, return all repositories unchanged
	if opts.ShowAll {
		return repos
	}

	// Filter to show only repos needing attention (not clean)
	filtered := make([]*models.Repository, 0, len(repos))
	for _, repo := range repos {
		if !IsClean(repo) {
			filtered = append(filtered, repo)
		}
	}

	return filtered
}
