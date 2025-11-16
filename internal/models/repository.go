package models

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Package-level color functions for status formatting.
//
//nolint:gochecknoglobals // These are immutable color functions, safe for concurrent use.
var (
	grayColor   = color.New(color.FgHiBlack, color.Bold).SprintFunc()
	yellowColor = color.New(color.FgYellow, color.Bold).SprintFunc()
	greenColor  = color.New(color.FgGreen, color.Bold).SprintFunc()
	redColor    = color.New(color.FgRed, color.Bold).SprintFunc()
)

// Repository represents a Git repository discovered during directory scanning.
type Repository struct {
	Path       string     // Absolute file system path to the repository directory
	Name       string     // Base name of the repository directory
	IsBare     bool       // Whether the repository is a bare repository
	IsSymlink  bool       // Whether the repository was reached via a symbolic link
	GitStatus  *GitStatus // Current Git status information (nil if error occurred)
	Error      error      // Error encountered during processing
	HasTimeout bool       // Whether Git operations timed out
}

var (
	errScanResultValidation = errors.New("scan result validation error")
	errRepositoryValidation = errors.New("repository validation error")
)

// Validate checks if the Repository meets all validation rules.
func (r *Repository) Validate() error {
	if r.Path == "" {
		return fmt.Errorf("path cannot be empty: %w", errRepositoryValidation)
	}
	if !filepath.IsAbs(r.Path) {
		return fmt.Errorf("path must be absolute: %w", errRepositoryValidation)
	}
	if r.Name == "" {
		return fmt.Errorf("name cannot be empty: %w", errRepositoryValidation)
	}
	if r.IsBare && r.GitStatus != nil && r.GitStatus.HasChanges {
		return fmt.Errorf("bare repository cannot have uncommitted changes: %w", errRepositoryValidation)
	}

	return nil
}

// GitStatus represents the Git status information for a repository.
type GitStatus struct {
	Branch     string // Current branch name or "DETACHED" if HEAD is detached
	IsDetached bool   // Whether HEAD is in detached state
	HasRemote  bool   // Whether repository has a remote configured
	Ahead      int    // Number of commits ahead of remote
	Behind     int    // Number of commits behind remote
	HasStashes bool   // Whether repository has stashed changes
	HasChanges bool   // Whether repository has uncommitted changes
	Error      string // Partial error message if some status info couldn't be retrieved
}

var errGitStatusValidation = errors.New("git status validation error")

// Validate checks if the GitStatus meets all validation rules.
func (g *GitStatus) Validate() error {
	if g.Branch == "" {
		return fmt.Errorf("branch cannot be empty: %w", errGitStatusValidation)
	}
	if g.IsDetached && g.Branch != "DETACHED" {
		return fmt.Errorf("detached HEAD must have branch = 'DETACHED': %w", errGitStatusValidation)
	}
	if !g.HasRemote && (g.Ahead != 0 || g.Behind != 0) {
		return fmt.Errorf("no remote but ahead/behind counts are non-zero: %w", errGitStatusValidation)
	}
	if g.Ahead < 0 || g.Behind < 0 {
		return fmt.Errorf("ahead/behind counts cannot be negative: %w", errGitStatusValidation)
	}

	return nil
}

// Format returns the formatted Git status string for display with colorization.
// Examples (with colors disabled):
//   - [[ main ]] - On main, in sync with remote, no changes
//   - [[ main | ↑2 ↓1 ]] - 2 commits ahead, 1 behind
//   - [[ develop | $ * ]] - Has stashes and uncommitted changes
//   - [[ DETACHED ]] - Detached HEAD state
//   - [[ main | ○ ]] - No remote configured
//   - [[ main ]] error - Partial error retrieving status
func (g *GitStatus) Format() string {
	var parts []string

	// Branch: gray for main/master, yellow otherwise
	if g.Branch == "main" || g.Branch == "master" {
		parts = append(parts, grayColor(g.Branch))
	} else {
		parts = append(parts, yellowColor(g.Branch))
	}

	// Ahead/Behind: green/red, or gray no-remote indicator
	if g.HasRemote {
		if g.Ahead > 0 {
			parts = append(parts, greenColor(fmt.Sprintf("↑%d", g.Ahead)))
		}
		if g.Behind > 0 {
			parts = append(parts, redColor(fmt.Sprintf("↓%d", g.Behind)))
		}
	} else {
		parts = append(parts, grayColor("○"))
	}

	// Stashes: red
	if g.HasStashes {
		parts = append(parts, redColor("$"))
	}

	// Uncommitted changes: red
	if g.HasChanges {
		parts = append(parts, redColor("*"))
	}

	// Build result with double gray brackets and separator
	var result string
	if len(parts) == 1 {
		// Only branch, no separator needed
		result = grayColor("[[") + " " + parts[0] + " " + grayColor("]]")
	} else {
		// Branch + status indicators, use separator
		separator := " " + grayColor("|") + " "
		statusParts := strings.Join(parts[1:], " ")
		result = grayColor("[[") + " " + parts[0] + separator + statusParts + " " + grayColor("]]")
	}

	// Append error indicator if present
	if g.Error != "" {
		result += " error"
	}

	return result
}

// TreeNode represents a node in the hierarchical tree structure.
type TreeNode struct {
	Repository   *Repository // The repository at this tree node
	Depth        int         // Depth level in the tree (0 = root)
	IsLast       bool        // Whether this is the last child of its parent
	Children     []*TreeNode // Child nodes (nested repositories)
	RelativePath string      // Path relative to scan root
}

var errTreeNodeValidation = errors.New("tree node validation error")

// Validate checks if the TreeNode meets all validation rules.
func (t *TreeNode) Validate() error {
	if t.Repository == nil {
		return fmt.Errorf("repository cannot be nil: %w", errTreeNodeValidation)
	}
	if t.Depth < 0 {
		return fmt.Errorf("depth cannot be negative: %d: %w", t.Depth, errTreeNodeValidation)
	}
	if t.RelativePath == "" {
		return fmt.Errorf("relative path cannot be empty: %s: %w", t.Repository.Path, errTreeNodeValidation)
	}

	return nil
}

// AddChild adds a child node to this tree node and sets the child's depth.
func (t *TreeNode) AddChild(child *TreeNode) {
	child.Depth = t.Depth + 1
	t.Children = append(t.Children, child)
}

// SortChildren sorts the children alphabetically by repository name
// and updates the IsLast flag for the last child.
func (t *TreeNode) SortChildren() {
	sort.Slice(t.Children, func(i, j int) bool {
		return t.Children[i].Repository.Name < t.Children[j].Repository.Name
	})
	// Update IsLast flags
	for i, child := range t.Children {
		child.IsLast = (i == len(t.Children)-1)
	}
}

// ScanResult represents the complete result of a directory scan operation.
type ScanResult struct {
	RootPath     string        // Absolute path where scan started
	Repositories []*Repository // All repositories found during scan
	Tree         *TreeNode     // Root node of the tree structure
	TotalScanned int           // Total number of directories scanned
	TotalRepos   int           // Total number of Git repositories found
	Errors       []error       // Collection of non-fatal errors
	Duration     time.Duration // Time taken to complete scan
}

// Validate checks if the ScanResult meets all validation rules.
func (s *ScanResult) Validate() error {
	if s.RootPath == "" {
		return fmt.Errorf("root path cannot be empty: %w", errScanResultValidation)
	}
	if !filepath.IsAbs(s.RootPath) {
		return fmt.Errorf("root path must be absolute: %s: %w", s.RootPath, errScanResultValidation)
	}
	if s.Repositories == nil {
		return fmt.Errorf("repositories slice cannot be nil: %w", errScanResultValidation)
	}
	if s.Tree == nil {
		return fmt.Errorf("tree cannot be nil: %w", errScanResultValidation)
	}
	if s.TotalRepos != len(s.Repositories) {
		return fmt.Errorf("total repos mismatch: %d != %d: %w", s.TotalRepos, len(s.Repositories), errScanResultValidation)
	}
	if s.TotalScanned < s.TotalRepos {
		return fmt.Errorf("total scanned < total repos: %d < %d: %w", s.TotalScanned, s.TotalRepos, errScanResultValidation)
	}
	if s.Duration < 0 {
		return fmt.Errorf("duration cannot be negative: %w", errScanResultValidation)
	}

	return nil
}

// HasErrors returns true if there are non-fatal errors in the scan result.
func (s *ScanResult) HasErrors() bool {
	return len(s.Errors) > 0
}

// SuccessRate returns the ratio of successful repositories to total repositories.
// Returns 1.0 if there are no repositories.
func (s *ScanResult) SuccessRate() float64 {
	if s.TotalRepos == 0 {
		return 1.0
	}

	reposWithErrors := 0
	for _, repo := range s.Repositories {
		if repo.Error != nil || repo.HasTimeout {
			reposWithErrors++
		}
	}

	return float64(s.TotalRepos-reposWithErrors) / float64(s.TotalRepos)
}
