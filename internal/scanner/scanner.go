package scanner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andreygrechin/gitree/internal/models"
)

// ScanOptions configures the directory scanning behavior.
type ScanOptions struct {
	RootPath string // Root directory to start scanning from
}

// IsGitRepository checks if a directory is a Git repository
// Returns (isRepo, isBare) where:
// - isRepo: true if directory contains a Git repository
// - isBare: true if it's a bare repository.
func IsGitRepository(path string) (isRepo, isBare bool) {
	// Check for regular repository (.git directory)
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		return true, false // regular repo
	}

	// Check for bare repository (HEAD, refs/, objects/ in root)
	headFile := filepath.Join(path, "HEAD")
	refsDir := filepath.Join(path, "refs")
	objsDir := filepath.Join(path, "objects")

	headExists := false
	if _, err := os.Stat(headFile); err == nil {
		headExists = true
	}

	refsExists := false
	if info, err := os.Stat(refsDir); err == nil && info.IsDir() {
		refsExists = true
	}

	objsExists := false
	if info, err := os.Stat(objsDir); err == nil && info.IsDir() {
		objsExists = true
	}

	// All three must exist for bare repo
	if headExists && refsExists && objsExists {
		return true, true // bare repo
	}

	return false, false
}

// scanner holds state during directory traversal.
type scanner struct {
	rootPath     string
	repositories []*models.Repository
	errors       []error
	visited      map[uint64]bool // Track visited inodes to prevent symlink loops
	dirCount     int
}

var errScanResultValidation = errors.New("scan result validation error")

// Scan recursively scans a directory tree for Git repositories.
func Scan(ctx context.Context, opts ScanOptions) (*models.ScanResult, error) {
	startTime := time.Now()

	// Validate root path exists
	info, err := os.Stat(opts.RootPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access root path: %w: %w", errScanResultValidation, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path %s is not a directory: %w", opts.RootPath, errScanResultValidation)
	}

	// Get absolute path
	absPath, err := filepath.Abs(opts.RootPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get absolute path: %w: %w", errScanResultValidation, err)
	}

	s := &scanner{
		rootPath:     absPath,
		repositories: make([]*models.Repository, 0),
		errors:       make([]error, 0),
		visited:      make(map[uint64]bool),
	}

	// Walk directory tree
	// Create closure to pass context to walkFunc
	walkFn := func(path string, d fs.DirEntry, err error) error {
		return s.walkFunc(ctx, path, d, err)
	}
	err = filepath.WalkDir(absPath, walkFn)
	if err != nil && !errors.Is(err, context.Canceled) {
		// If it's not a context cancellation, it's a fatal error
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}

	// Build tree structure from flat repository list
	tree := s.buildTree()

	result := &models.ScanResult{
		RootPath:     absPath,
		Repositories: s.repositories,
		Tree:         tree,
		TotalScanned: s.dirCount,
		TotalRepos:   len(s.repositories),
		Errors:       s.errors,
		Duration:     time.Since(startTime),
	}

	return result, nil
}

var errPermissionDenied = errors.New("permission denied")

// walkFunc is called for each file/directory during traversal.
func (s *scanner) walkFunc(ctx context.Context, path string, d fs.DirEntry, err error) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Handle permission errors (non-fatal)
	if err != nil {
		if os.IsPermission(err) {
			s.errors = append(s.errors, fmt.Errorf("permission denied: %s: %w", path, errPermissionDenied))

			return fs.SkipDir // Skip this directory but continue scanning
		}

		return err
	}

	// Only process directories
	if !d.IsDir() {
		return nil
	}

	s.dirCount++

	// Check for symlink loops using inode tracking
	shouldVisit, isSymlink, err := s.shouldVisit(path)
	if err != nil {
		s.errors = append(s.errors, fmt.Errorf("error checking path %s: %w", path, err))

		return fs.SkipDir
	}
	if !shouldVisit {
		return fs.SkipDir // Already visited or symlink loop
	}

	// Check if this directory is a Git repository
	isRepo, isBare := IsGitRepository(path)
	if isRepo {
		repo := &models.Repository{
			Path:      path,
			Name:      filepath.Base(path),
			IsBare:    isBare,
			IsSymlink: isSymlink,
		}

		s.repositories = append(s.repositories, repo)

		// Skip traversing into repository contents (FR-018)
		// We found a repo, so we don't need to look inside it for more repos
		return fs.SkipDir
	}

	return nil
}

// shouldVisit checks if a path should be visited (handles symlink loops)
// Returns (shouldVisit, isSymlink, error).
func (s *scanner) shouldVisit(path string) (shouldVisit, isSymlink bool, err error) {
	// Get file info without following symlinks
	info, err := os.Lstat(path)
	if err != nil {
		return false, false, err
	}

	// Check if it's a symlink
	isSymlink = info.Mode()&os.ModeSymlink != 0
	var actualPath string

	if isSymlink {
		// Follow symlink to get real path
		actualPath, err = filepath.EvalSymlinks(path)
		if err != nil {
			// Broken symlink
			return false, false, err
		}

		// Get info of the target
		info, err = os.Stat(actualPath)
		if err != nil {
			return false, false, err
		}

		// If symlink target is not a directory, skip
		if !info.IsDir() {
			return false, false, nil
		}
	}

	// Get inode to track visited paths
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		// Can't get inode (might be on Windows), just visit it
		return true, isSymlink, nil
	}

	inode := stat.Ino

	// Check if already visited
	if s.visited[inode] {
		return false, isSymlink, nil // Already visited, skip
	}

	// Mark as visited
	s.visited[inode] = true

	return true, isSymlink, nil
}

// buildTree creates a TreeNode structure from flat repository list.
func (s *scanner) buildTree() *models.TreeNode {
	if len(s.repositories) == 0 {
		// No repositories found - create empty root node
		return &models.TreeNode{
			Repository: &models.Repository{
				Path: s.rootPath,
				Name: filepath.Base(s.rootPath),
			},
			Depth:        0,
			IsLast:       true,
			Children:     []*models.TreeNode{},
			RelativePath: ".",
		}
	}

	// For now, create a simple flat tree
	// This will be enhanced in the tree formatter phase
	// Each repository becomes a root-level node

	root := &models.TreeNode{
		Repository: &models.Repository{
			Path: s.rootPath,
			Name: filepath.Base(s.rootPath),
		},
		Depth:        0,
		IsLast:       false,
		Children:     make([]*models.TreeNode, 0),
		RelativePath: ".",
	}

	for _, repo := range s.repositories {
		relPath, err := filepath.Rel(s.rootPath, repo.Path)
		if err != nil {
			relPath = repo.Path
		}

		node := &models.TreeNode{
			Repository:   repo,
			Depth:        1,
			IsLast:       false,
			Children:     []*models.TreeNode{},
			RelativePath: relPath,
		}

		root.Children = append(root.Children, node)
	}

	// Sort children and mark last
	root.SortChildren()

	return root
}
