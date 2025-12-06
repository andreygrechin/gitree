package models

import "testing"

// FuzzGitStatusFormat tests GitStatus.Format() with random inputs.
// It verifies the function never panics regardless of input values.
func FuzzGitStatusFormat(f *testing.F) {
	// Seed corpus with representative test cases
	f.Add("main", 0, 0, false, false, false, false, "", "")
	f.Add("master", 0, 0, true, false, false, false, "", "")
	f.Add("feature/test", 5, 3, true, true, true, false, "", "")
	f.Add("DETACHED", 0, 0, false, false, false, true, "", "")
	f.Add("", 0, 0, false, false, false, false, "", "")
	f.Add("N/A", 0, 0, false, false, false, false, "error message", "")
	f.Add("main", 100, 100, true, true, true, false, "", "fetch failed")
	f.Add("main", -1, -1, false, false, false, false, "", "")
	f.Add("branch", 999999, 999999, true, false, false, false, "", "")

	f.Fuzz(func(_ *testing.T, branch string, ahead, behind int,
		hasRemote, hasStashes, hasChanges, isDetached bool,
		errStr, fetchErr string,
	) {
		status := &GitStatus{
			Branch:     branch,
			Ahead:      ahead,
			Behind:     behind,
			HasRemote:  hasRemote,
			HasStashes: hasStashes,
			HasChanges: hasChanges,
			IsDetached: isDetached,
			Error:      errStr,
			FetchError: fetchErr,
		}

		// Format should never panic regardless of input
		_ = status.Format()
	})
}

// FuzzGitStatusValidate tests GitStatus.Validate() with random inputs.
// It verifies the function returns an error or nil without panicking.
func FuzzGitStatusValidate(f *testing.F) {
	// Seed corpus with valid and invalid states
	f.Add("main", 0, 0, true, false)
	f.Add("master", 0, 0, true, false)
	f.Add("DETACHED", 0, 0, false, true)
	f.Add("feature", 5, 3, true, false)
	f.Add("", 0, 0, false, false)            // Invalid: empty branch
	f.Add("feature", -1, 0, true, false)     // Invalid: negative ahead
	f.Add("feature", 0, -1, true, false)     // Invalid: negative behind
	f.Add("main", 5, 0, false, false)        // Invalid: ahead without remote
	f.Add("wrong-branch", 0, 0, false, true) // Invalid: detached with wrong branch
	f.Add("DETACHED", 0, 0, true, true)      // Valid: detached with remote

	f.Fuzz(func(_ *testing.T, branch string, ahead, behind int,
		hasRemote, isDetached bool,
	) {
		status := &GitStatus{
			Branch:     branch,
			Ahead:      ahead,
			Behind:     behind,
			HasRemote:  hasRemote,
			IsDetached: isDetached,
		}

		// Validate should return error or nil, never panic
		_ = status.Validate()
	})
}
