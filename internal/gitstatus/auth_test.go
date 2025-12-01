package gitstatus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// T_A001: Test isHTTPSURL correctly identifies HTTPS URLs.
func TestIsHTTPSURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"HTTPS URL", "https://github.com/user/repo.git", true},
		{"HTTPS URL without .git", "https://github.com/user/repo", true},
		{"HTTP URL", "http://github.com/user/repo.git", false},
		{"SSH URL with git@", "git@github.com:user/repo.git", false},
		{"SSH URL with ssh://", "ssh://git@github.com/user/repo.git", false},
		{"File URL", "file:///path/to/repo", false},
		{"HTTPS with uppercase", "HTTPS://github.com/user/repo.git", true},
		{"Local path", "/path/to/repo", false},
		{"Git protocol", "git://github.com/user/repo.git", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHTTPSURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// T_A002: Test getAuthForURL returns nil for SSH URLs.
func TestGetAuthForURL_SSHReturnsNil(t *testing.T) {
	ctx := context.Background()

	sshURLs := []string{
		"git@github.com:user/repo.git",
		"ssh://git@github.com/user/repo.git",
	}

	for _, url := range sshURLs {
		t.Run(url, func(t *testing.T) {
			auth := getAuthForURL(ctx, url, false)
			assert.Nil(t, auth, "SSH URL should return nil auth")
		})
	}
}

// T_A003: Test getAuthForURL returns nil for HTTP URLs.
func TestGetAuthForURL_HTTPReturnsNil(t *testing.T) {
	ctx := context.Background()

	auth := getAuthForURL(ctx, "http://github.com/user/repo.git", false)
	assert.Nil(t, auth, "HTTP URL should return nil auth")
}

// T_A004: Test getAuthForURL returns nil for file paths.
func TestGetAuthForURL_FilePathReturnsNil(t *testing.T) {
	ctx := context.Background()

	paths := []string{
		"/path/to/repo",
		"file:///path/to/repo",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			auth := getAuthForURL(ctx, path, false)
			assert.Nil(t, auth, "File path should return nil auth")
		})
	}
}

// T_A005: Test getGitCredentials with invalid URL.
func TestGetGitCredentials_InvalidURL(t *testing.T) {
	ctx := context.Background()

	creds, err := getGitCredentials(ctx, "://invalid", false)

	assert.Nil(t, creds)
	assert.Error(t, err)
}

// T_A006: Test getGitCredentials respects context cancellation.
func TestGetGitCredentials_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	creds, err := getGitCredentials(ctx, "https://github.com/user/repo.git", false)

	// Should fail due to context cancellation
	assert.Nil(t, creds)
	assert.Error(t, err)
}

// T_A007: Test getGitCredentials respects timeout.
func TestGetGitCredentials_RespectsTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a moment to ensure timeout fires
	time.Sleep(10 * time.Millisecond)

	creds, err := getGitCredentials(ctx, "https://github.com/user/repo.git", false)

	// Should fail due to context timeout
	assert.Nil(t, creds)
	assert.Error(t, err)
}

// T_A008: Test getAuthForURL with debug mode doesn't panic.
func TestGetAuthForURL_DebugMode(t *testing.T) {
	ctx := context.Background()

	// Should not panic with debug enabled
	assert.NotPanics(t, func() {
		_ = getAuthForURL(ctx, "git@github.com:user/repo.git", true)
		_ = getAuthForURL(ctx, "https://github.com/user/repo.git", true)
		_ = getAuthForURL(ctx, "/path/to/repo", true)
	})
}
