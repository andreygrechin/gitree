package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVersionFlag_OutputFormat verifies that version flag displays all three required fields.
func TestVersionFlag_OutputFormat(t *testing.T) {
	// Save original version variables
	origVersion := version
	origCommit := commit
	origBuildTime := buildTime

	// Set test values
	version = "v1.2.3"
	commit = "a1b2c3d"
	buildTime = "2025-11-17T14:30:00Z"

	// Restore after test
	defer func() {
		version = origVersion
		commit = origCommit
		buildTime = origBuildTime
	}()

	// Capture output
	output := captureVersionOutput(t)

	// Verify all three fields are present
	assert.Contains(t, output, "gitree version v1.2.3", "Version should be displayed")
	assert.Contains(t, output, "commit: a1b2c3d", "Commit hash should be displayed")
	assert.Contains(t, output, "built:  2025-11-17T14:30:00Z", "Build time should be displayed")

	// Verify format structure
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 3, "Output should have exactly 3 lines")
	assert.True(t, strings.HasPrefix(lines[0], "gitree version"), "First line should start with 'gitree version'")
	assert.True(t, strings.HasPrefix(lines[1], "  commit:"), "Second line should be indented and show commit")
	assert.True(t, strings.HasPrefix(lines[2], "  built:"), "Third line should be indented and show build time")
}

// TestVersionFlag_DefaultValues verifies version flag works with default values (dev build).
func TestVersionFlag_DefaultValues(t *testing.T) {
	// Save original version variables
	origVersion := version
	origCommit := commit
	origBuildTime := buildTime

	// Set default values (as they appear in dev builds)
	version = "dev"
	commit = "none"
	buildTime = "unknown"

	// Restore after test
	defer func() {
		version = origVersion
		commit = origCommit
		buildTime = origBuildTime
	}()

	// Capture output
	output := captureVersionOutput(t)

	// Verify default values are displayed
	assert.Contains(t, output, "gitree version dev", "Should show 'dev' for version")
	assert.Contains(t, output, "commit: none", "Should show 'none' for commit")
	assert.Contains(t, output, "built:  unknown", "Should show 'unknown' for build time")
}

// TestVersionFlag_ExecutionTime verifies version flag executes quickly (<100ms per SC-001).
func TestVersionFlag_ExecutionTime(t *testing.T) {
	start := time.Now()

	// Capture output (this triggers version display)
	_ = captureVersionOutput(t)

	duration := time.Since(start)

	// Verify execution time is under 100ms (SC-001)
	assert.Less(t, duration, 100*time.Millisecond, "Version flag should execute in less than 100ms")
}

// TestVersionFlag_Precedence verifies version flag exits immediately without scanning.
func TestVersionFlag_Precedence(t *testing.T) {
	// This test verifies that --version flag takes precedence and exits
	// before any repository scanning occurs.

	// Use helper to capture output (includes exit mocking)
	output := captureVersionOutput(t)

	// Verify version output was generated
	assert.Contains(t, output, "gitree version", "Should display version information")

	// Note: We can't directly verify "no scanning occurred" in this unit test,
	// but the implementation uses PreRunE to exit early before runGitree is called.
	// Integration tests will verify this behavior more thoroughly.
}

// Helper function to capture version display output.
func captureVersionOutput(t *testing.T) string {
	t.Helper()

	// Reset command for test
	resetRootCommand()

	// Mock exitAfterVersion to not actually exit in tests
	origExit := exitAfterVersion
	exitAfterVersion = func() error { return nil }
	defer func() { exitAfterVersion = origExit }()

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Set the args directly on the command
	rootCmd.SetArgs([]string{"--version"})

	// Execute command
	err := rootCmd.Execute()
	require.NoError(t, err, "Version command should not error")

	return buf.String()
}

// Helper function to reset root command for testing.
func resetRootCommand() {
	// Reset flags to default values
	versionFlag = false
	noColorFlag = false
	allFlag = false

	// Reset command args
	rootCmd.SetArgs([]string{})
}
