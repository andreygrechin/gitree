package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/andreygrechin/gitree/internal/gitstatus"
	"github.com/andreygrechin/gitree/internal/models"
	"github.com/andreygrechin/gitree/internal/scanner"
	"github.com/andreygrechin/gitree/internal/tree"
	"github.com/briandowns/spinner"
)

const (
	defaultTimeout        = 10 * time.Second
	maxConcurrentRequests = 10
	spinnerDelay          = 100 * time.Millisecond
	spinnerChar           = 11
	defaultContextTimeout = 5 * time.Minute
)

func main() {
	var exitCode int
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize spinner

	s := spinner.New(spinner.CharSets[spinnerChar], spinnerDelay)
	s.Suffix = " Scanning repositories..."
	s.Writer = os.Stderr
	s.Start()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer func() {
		cancel()
		os.Exit(exitCode)
	}()

	// Scan for repositories
	scanOpts := scanner.ScanOptions{
		RootPath: cwd,
	}
	scanResult, err := scanner.Scan(ctx, scanOpts)
	if err != nil {
		s.Stop()
		fmt.Fprintf(os.Stderr, "Error: Failed to scan directory: %v\n", err)
		exitCode = 1

		return
	}

	// Check if any repositories were found
	if len(scanResult.Repositories) == 0 {
		s.Stop()
		_, _ = fmt.Fprintln(os.Stdout, "No Git repositories found in this directory.")
		exitCode = 0

		return
	}

	// Update spinner message
	s.Suffix = fmt.Sprintf(" Extracting Git status for %d repositories...", len(scanResult.Repositories))

	// Create map of repositories for batch processing
	repoMap := make(map[string]*models.Repository)
	for _, repo := range scanResult.Repositories {
		repoMap[repo.Path] = repo
	}

	// Extract Git status concurrently
	statusOpts := &gitstatus.ExtractOptions{
		Timeout:        defaultTimeout,
		MaxConcurrency: maxConcurrentRequests,
	}
	statuses, err := gitstatus.ExtractBatch(ctx, repoMap, statusOpts)
	if err != nil {
		s.Stop()
		fmt.Fprintf(os.Stderr, "Warning: Some repositories failed status extraction: %v\n", err)
		// Continue anyway with partial results
	}

	// Populate repositories with status
	for path, status := range statuses {
		if repo, exists := repoMap[path]; exists {
			repo.GitStatus = status
		}
	}

	// Build tree structure
	root := tree.Build(cwd, scanResult.Repositories, nil)

	// Stop spinner before output
	s.Stop()

	// Format and print tree
	output := tree.Format(root, nil)
	_, _ = fmt.Fprint(os.Stdout, output)

	// Exit successfully
	exitCode = 0
}
