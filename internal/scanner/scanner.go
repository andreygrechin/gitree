//go:build !windows

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
	Debug    bool   // Enable debug output for scanning operations
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

// debugPrintf formats the message using fmt.Sprintf, adds a "DEBUG: " prefix, and outputs it to stderr if debug is enabled.
func debugPrintf(debug bool, format string, args ...any) {
	if !debug {
		return
	}
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "DEBUG: %s\n", message)
}

// scanner holds state during directory traversal.
type scanner struct {
	rootPath     string
	opts         ScanOptions // Store full options instead of just rootPath
	repositories []*models.Repository
	errors       []error
	visited      map[uint64]bool // Track visited inodes to prevent symlink loops
	dirCount     int
}

var errScanOptValidation = errors.New("scan options validation error")

// Scan recursively scans a directory tree for Git repositories.
func Scan(ctx context.Context, opts ScanOptions) (*models.ScanResult, error) {
	startTime := time.Now()

	// Validate root path exists
	info, err := os.Stat(opts.RootPath)
	if err != nil {
		return nil, fmt.Errorf("invalid scan options: cannot access root path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path %s is not a directory: %w", opts.RootPath, errScanOptValidation)
	}

	// Get absolute path
	absPath, err := filepath.Abs(opts.RootPath)
	if err != nil {
		return nil, fmt.Errorf("invalid scan options: cannot get absolute path: %w", err)
	}

	s := &scanner{
		rootPath:     absPath,
		opts:         opts,
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

	result := &models.ScanResult{
		RootPath:     absPath,
		Repositories: s.repositories,
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
			debugPrintf(s.opts.Debug, "Skipping %s: permission denied", path)
			s.errors = append(s.errors, fmt.Errorf("permission denied: %s: %w", path, errPermissionDenied))

			return fs.SkipDir // Skip this directory but continue scanning
		}

		return err
	}

	// Only process directories
	if !d.IsDir() {
		return nil
	}

	debugPrintf(s.opts.Debug, "Entering directory: %s", path)

	s.dirCount++

	// Check for symlink loops using inode tracking
	shouldVisit, isSymlink, err := s.shouldVisit(path)
	if err != nil {
		debugPrintf(s.opts.Debug, "Skipping %s: %v", path, err)
		s.errors = append(s.errors, fmt.Errorf("error checking path %s: %w", path, err))

		return fs.SkipDir
	}
	if !shouldVisit {
		debugPrintf(s.opts.Debug, "Skipping %s: already visited (symlink loop)", path)

		return fs.SkipDir // Already visited or symlink loop
	}

	// Check if this directory is a Git repository
	isRepo, isBare := IsGitRepository(path)
	if isRepo {
		repoType := "regular"
		if isBare {
			repoType = "bare"
		}
		debugPrintf(s.opts.Debug, "Found git repository: %s (%s)", path, repoType)

		repo := &models.Repository{
			Path:      path,
			Name:      filepath.Base(path),
			IsBare:    isBare,
			IsSymlink: isSymlink,
		}

		s.repositories = append(s.repositories, repo)

		// Skip traversing into repository contents (FR-018)
		// We found a repo, so we don't need to look inside it for more repos
		debugPrintf(s.opts.Debug, "Skipping %s: inside git repository", path)

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
