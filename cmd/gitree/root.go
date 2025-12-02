package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/andreygrechin/gitree/internal/cli"
	"github.com/andreygrechin/gitree/internal/gitstatus"
	"github.com/andreygrechin/gitree/internal/models"
	"github.com/andreygrechin/gitree/internal/scanner"
	"github.com/andreygrechin/gitree/internal/tree"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	defaultTimeout        = 10 * time.Second
	defaultMaxConcurrent  = 50
	spinnerDelay          = 100 * time.Millisecond
	spinnerChar           = 11
	defaultContextTimeout = 5 * time.Minute
)

var errInvalidFlags = errors.New("invalid flag value")

//nolint:gochecknoglobals // CLI flags and root command
var (
	// Flags.
	versionFlag       bool
	noColorFlag       bool
	allFlag           bool
	debugFlag         bool
	noFetchFlag       bool
	maxConcurrentFlag int

	// Root command.
	rootCmd = &cobra.Command{
		Use:   "gitree [directory]",
		Short: "Recursively scan directories for Git repositories and display them in a tree structure",
		Long: `gitree scans the specified directory (or current directory if not provided) and its
subdirectories for Git repositories, displays them in a tree structure with status information.

By default, gitree fetches from origin remote before calculating ahead/behind
counts. Use --no-fetch to skip fetching and use local refs only.

By default, only repositories needing attention are shown (uncommitted changes,
non-main/master branches, ahead/behind remote, stashes, or no remote tracking).
Use --all to show all repositories including clean ones.`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runGitree,
	}

	// exitAfterVersion signals that we should exit after showing version.
	// This is a separate function to make it testable.
	exitAfterVersion = func() error {
		os.Exit(0)

		return nil // Unreachable but required for testability
	}
)

func init() { //nolint:gochecknoinits // Cobra CLI initialization
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVar(&noColorFlag, "no-color", false, "Disable color output")
	rootCmd.Flags().BoolVarP(&allFlag, "all", "a", false,
		"Show all repositories including clean ones (default shows only repos needing attention)")
	rootCmd.Flags().BoolVar(&debugFlag, "debug", false, "Enable debug output")
	rootCmd.Flags().BoolVar(&noFetchFlag, "no-fetch", false,
		"Skip fetching from remote (use local refs only)")
	rootCmd.Flags().IntVarP(&maxConcurrentFlag, "max-concurrent", "c", defaultMaxConcurrent,
		"Maximum concurrent git operations")

	// Set PersistentPreRun to handle global flags (color suppression)
	rootCmd.PersistentPreRun = handleGlobalFlags

	// Set PreRunE to handle version flag with precedence
	rootCmd.PreRunE = handleVersionFlag
}

// handleGlobalFlags handles global flags that affect all commands.
func handleGlobalFlags(_ *cobra.Command, _ []string) {
	// Handle color suppression (--no-color flag or NO_COLOR environment variable)
	if noColorFlag || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}

// handleVersionFlag handles the --version flag with precedence over other operations and validates flag values.
func handleVersionFlag(cmd *cobra.Command, _ []string) error {
	if versionFlag {
		displayVersion(cmd)
		// Exit after displaying version (prevents further execution)
		// Note: In tests, this will be caught by cmd.Execute() returning nil
		return exitAfterVersion()
	}

	if maxConcurrentFlag < 1 {
		return fmt.Errorf("%w: --max-concurrent must be at least 1, got %d", errInvalidFlags, maxConcurrentFlag)
	}

	return nil
}

// displayVersion formats and displays version information.
func displayVersion(cmd *cobra.Command) {
	output := formatVersion(version, commit, buildTime)
	cmd.Println(output)
}

// formatVersion returns the formatted version string.
func formatVersion(ver, cmt, btime string) string {
	return fmt.Sprintf("gitree version %s\n  commit: %s\n  built:  %s", ver, cmt, btime)
}

func runGitree(_ *cobra.Command, args []string) error { //nolint:gocognit // Main command logic
	// Determine target directory
	var targetDir string
	if len(args) > 0 {
		targetDir = args[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to get current directory: %w", err)
		}
		targetDir = cwd
	}

	// Initialize spinner
	s := spinner.New(spinner.CharSets[spinnerChar], spinnerDelay)
	s.Suffix = " Scanning repositories..."
	s.Writer = os.Stderr
	// Only start spinner if debug is disabled
	if !debugFlag {
		s.Start()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	// Scan for repositories
	scanOpts := scanner.ScanOptions{
		RootPath: targetDir,
		Debug:    debugFlag,
	}
	scanResult, err := scanner.Scan(ctx, scanOpts)
	if err != nil {
		if !debugFlag {
			s.Stop()
		}

		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Validate scan result
	if valErr := scanResult.Validate(); valErr != nil {
		logValidationWarning("ScanResult validation failed", valErr)
	}

	// Check if any repositories were found
	if len(scanResult.Repositories) == 0 {
		if !debugFlag {
			s.Stop()
		}
		_, _ = fmt.Fprintln(os.Stdout, "No Git repositories found in this directory.")
		// Still print summary even when no repos found
		printSummary(scanResult, &models.BatchResult{})

		return nil
	}

	// Update spinner message
	if noFetchFlag {
		s.Suffix = fmt.Sprintf(" Extracting Git status for %d repositories...", len(scanResult.Repositories))
	} else {
		s.Suffix = fmt.Sprintf(" Fetching and extracting Git status for %d repositories...", len(scanResult.Repositories))
	}

	// Create map of repositories for batch processing
	repoMap := make(map[string]*models.Repository)
	for _, repo := range scanResult.Repositories {
		repoMap[repo.Path] = repo
	}

	// Extract Git status concurrently (with fetch if enabled)
	statusOpts := &gitstatus.ExtractOptions{
		Timeout:        defaultTimeout,
		MaxConcurrency: maxConcurrentFlag,
		Debug:          debugFlag,
		Fetch:          !noFetchFlag,
	}
	batchResult := gitstatus.ExtractBatch(ctx, repoMap, statusOpts)

	// Populate repositories with status
	for path, status := range batchResult.Statuses {
		if repo, exists := repoMap[path]; exists {
			repo.GitStatus = status
		}
	}

	// Validate repositories and their git status
	for _, repo := range scanResult.Repositories {
		if valErr := repo.Validate(); valErr != nil {
			logValidationWarning("Repository validation failed for "+repo.Path, valErr)
		}
		if repo.GitStatus != nil {
			if valErr := repo.GitStatus.Validate(); valErr != nil {
				logValidationWarning("GitStatus validation failed for "+repo.Path, valErr)
			}
		}
	}

	// Filter repositories based on --all flag
	filterOpts := cli.FilterOptions{ShowAll: allFlag}
	filteredRepos := cli.FilterRepositories(scanResult.Repositories, filterOpts)

	// Check if all repos were filtered out (all clean in default mode)
	if len(filteredRepos) == 0 && !allFlag {
		if !debugFlag {
			s.Stop()
		}
		_, _ = fmt.Fprintln(os.Stdout, "All repositories are in clean state (on main/master, in sync with remote, no changes).")
		_, _ = fmt.Fprintln(os.Stdout, "Use --all flag to show all repositories including clean ones.")
		// Still print summary even when all repos filtered
		printSummary(scanResult, batchResult)

		return nil
	}

	// Build tree structure with filtered repositories
	root := tree.Build(targetDir, filteredRepos, nil)

	// Validate tree structure
	if valErr := validateTree(root); valErr != nil {
		logValidationWarning("Tree validation failed", valErr)
	}

	// Stop spinner before output
	if !debugFlag {
		s.Stop()
	}

	// Format and print tree
	output := tree.Format(root, nil)
	_, _ = fmt.Fprint(os.Stdout, output)

	// Print summary statistics
	printSummary(scanResult, batchResult)

	return nil
}

// logValidationWarning logs a validation warning to stderr if debug mode is enabled.
func logValidationWarning(msg string, err error) {
	if debugFlag {
		_, _ = fmt.Fprintf(os.Stderr, "WARN: %s: %v\n", msg, err)
	}
}

// validateTree recursively validates all nodes in the tree.
func validateTree(node *models.TreeNode) error {
	if node == nil {
		return nil
	}
	if err := node.Validate(); err != nil {
		return err
	}
	for _, child := range node.Children {
		if err := validateTree(child); err != nil {
			return err
		}
	}

	return nil
}

// printSummary outputs statistics about the scan and fetch operations.
func printSummary(scanResult *models.ScanResult, batchResult *models.BatchResult) {
	_, _ = fmt.Fprintln(os.Stderr)
	_, _ = fmt.Fprintf(os.Stderr, "Scanned: %d folders\n", scanResult.TotalScanned)
	_, _ = fmt.Fprintf(os.Stderr, "Found: %d repositories\n", scanResult.TotalRepos)

	if batchResult.FetchStats != nil && batchResult.FetchStats.TotalAttempted > 0 {
		stats := batchResult.FetchStats
		_, _ = fmt.Fprintf(os.Stderr, "Fetch: %d attempted, %d successful, %d skipped, %d failed\n",
			stats.TotalAttempted, stats.Successful, stats.Skipped, stats.Failed)

		// Print failed repos if any
		if stats.Failed > 0 {
			_, _ = fmt.Fprintln(os.Stderr, "\nFetch failures:")
			for _, path := range stats.FailedRepos {
				_, _ = fmt.Fprintf(os.Stderr, "  - %s\n", path)
			}
		}
	}
}
