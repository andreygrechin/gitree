package gitstatus

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	credentialTimeout = 10 * time.Second
	keyValueParts     = 2 // Expected number of parts when splitting key=value.
)

var errNoCredentials = errors.New("no credentials available")

// gitCredentials holds credentials obtained from git credential helper.
type gitCredentials struct {
	Protocol string
	Host     string
	Path     string
	Username string
	Password string
}

// isHTTPSURL returns true if the URL is an HTTPS URL requiring authentication.
func isHTTPSURL(remoteURL string) bool {
	parsed, err := url.Parse(remoteURL)
	if err != nil {
		// Fallback: check if it starts with https://
		return strings.HasPrefix(strings.ToLower(remoteURL), "https://")
	}

	return strings.EqualFold(parsed.Scheme, "https")
}

// getAuthForURL returns the appropriate authentication method for a remote URL.
// Returns nil if no authentication is needed (SSH) or credentials cannot be obtained.
func getAuthForURL(ctx context.Context, remoteURL string, debug bool) transport.AuthMethod {
	if !isHTTPSURL(remoteURL) {
		// SSH URLs don't need explicit auth - go-git uses ssh-agent automatically
		if debug {
			debugPrintf("URL %s is not HTTPS, skipping credential lookup", remoteURL)
		}

		return nil
	}

	// Try to get credentials from Git credential helper
	creds, err := getGitCredentials(ctx, remoteURL, debug)
	if err != nil {
		if debug {
			if errors.Is(err, errNoCredentials) {
				debugPrintf("No credentials found for %s", remoteURL)
			} else {
				debugPrintf("Failed to get credentials for %s: %v", remoteURL, err)
			}
		}

		return nil
	}

	if debug {
		debugPrintf("Using credentials for %s (username: %s)", remoteURL, creds.Username)
	}

	return &http.BasicAuth{
		Username: creds.Username,
		Password: creds.Password,
	}
}

// getGitCredentials invokes the git credential helper to obtain credentials.
func getGitCredentials(ctx context.Context, remoteURL string, debug bool) (*gitCredentials, error) {
	// Parse the URL to extract components
	parsed, err := url.Parse(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Build the credential input
	var input bytes.Buffer
	input.WriteString(fmt.Sprintf("protocol=%s\n", parsed.Scheme))
	input.WriteString(fmt.Sprintf("host=%s\n", parsed.Host))

	if parsed.Path != "" {
		// Remove leading slash and .git suffix for path
		path := strings.TrimPrefix(parsed.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		if path != "" {
			input.WriteString(fmt.Sprintf("path=%s\n", path))
		}
	}

	input.WriteString("\n") // Empty line signals end of input

	// Create context with timeout for credential helper
	credCtx, cancel := context.WithTimeout(ctx, credentialTimeout)
	defer cancel()

	// Execute git credential fill
	cmd := exec.CommandContext(credCtx, "git", "credential", "fill")
	cmd.Stdin = &input

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if debug {
		debugPrintf("Running git credential fill for protocol=%s host=%s", parsed.Scheme, parsed.Host)
	}

	err = cmd.Run()
	if err != nil {
		if debug {
			debugPrintf("git credential fill failed: %v, stderr: %s", err, stderr.String())
		}

		return nil, fmt.Errorf("git credential fill failed: %w", err)
	}

	// Parse the output
	creds := &gitCredentials{}
	scanner := bufio.NewScanner(&stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", keyValueParts)
		if len(parts) != keyValueParts {
			continue
		}

		key, value := parts[0], parts[1]

		switch key {
		case "protocol":
			creds.Protocol = value
		case "host":
			creds.Host = value
		case "path":
			creds.Path = value
		case "username":
			creds.Username = value
		case "password":
			creds.Password = value
		}
	}

	if creds.Username == "" || creds.Password == "" {
		if debug {
			debugPrintf("Credentials incomplete: username=%t password=%t",
				creds.Username != "", creds.Password != "")
		}

		return nil, errNoCredentials
	}

	return creds, nil
}
