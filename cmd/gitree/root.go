package main

import (
	"context"
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
	maxConcurrentRequests = 10
	spinnerDelay          = 100 * time.Millisecond
	spinnerChar           = 11
	defaultContextTimeout = 5 * time.Minute
)

var (
	// Flags.
	versionFlag bool
	noColorFlag bool
	allFlag     bool

	// Root command.
	rootCmd = &cobra.Command{
		Use:   "gitree",
		Short: "Recursively scan directories for Git repositories and display them in a tree structure",
		Long: `gitree scans the current directory and its subdirectories for Git repositories,
displays them in a tree structure with status information.

By default, only repositories needing attention are shown (uncommitted changes,
non-main/master branches, ahead/behind remote, stashes, or no remote tracking).
Use --all to show all repositories including clean ones.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runGitree,
	}
)

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVar(&noColorFlag, "no-color", false, "Disable color output")
	rootCmd.Flags().BoolVar(&allFlag, "all", false, "Show all repositories including clean ones (default shows only repos needing attention)")

	// Set PersistentPreRun to handle global flags (color suppression)
	rootCmd.PersistentPreRun = handleGlobalFlags

	// Set PreRunE to handle version flag with precedence
	rootCmd.PreRunE = handleVersionFlag
}

// handleGlobalFlags handles global flags that affect all commands.
func handleGlobalFlags(cmd *cobra.Command, args []string) {
	// Handle color suppression (--no-color flag or NO_COLOR environment variable)
	if noColorFlag || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}

// handleVersionFlag handles the --version flag with precedence over other operations.
func handleVersionFlag(cmd *cobra.Command, args []string) error {
	if versionFlag {
		displayVersion(cmd)
		// Exit after displaying version (prevents further execution)
		// Note: In tests, this will be caught by cmd.Execute() returning nil
		return exitAfterVersion()
	}

	return nil
}

// exitAfterVersion signals that we should exit after showing version.
// This is a separate function to make it testable.
var exitAfterVersion = func() error {
	os.Exit(0)

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

func runGitree(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get current directory: %w", err)
	}

	// Initialize spinner
	s := spinner.New(spinner.CharSets[spinnerChar], spinnerDelay)
	s.Suffix = " Scanning repositories..."
	s.Writer = os.Stderr
	s.Start()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	// Scan for repositories
	scanOpts := scanner.ScanOptions{
		RootPath: cwd,
	}
	scanResult, err := scanner.Scan(ctx, scanOpts)
	if err != nil {
		s.Stop()

		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Check if any repositories were found
	if len(scanResult.Repositories) == 0 {
		s.Stop()
		fmt.Fprintln(os.Stdout, "No Git repositories found in this directory.")

		return nil
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

	// Filter repositories based on --all flag
	filterOpts := cli.FilterOptions{ShowAll: allFlag}
	filteredRepos := cli.FilterRepositories(scanResult.Repositories, filterOpts)

	// Check if all repos were filtered out (all clean in default mode)
	if len(filteredRepos) == 0 && !allFlag {
		s.Stop()
		fmt.Fprintln(os.Stdout, "All repositories are in clean state (on main/master, in sync with remote, no changes).")
		fmt.Fprintln(os.Stdout, "Use --all flag to show all repositories including clean ones.")

		return nil
	}

	// Build tree structure with filtered repositories
	root := tree.Build(cwd, filteredRepos, nil)

	// Stop spinner before output
	s.Stop()

	// Format and print tree
	output := tree.Format(root, nil)
	fmt.Fprint(os.Stdout, output)

	return nil
}
