# ANSI Escape Sequence Validation Contract

## Purpose

Define the exact validation criteria for SC-002: "Output with no-color flag contains zero ANSI escape sequences."

## ANSI Escape Sequence Pattern

**Regex Pattern**: `\x1b\[[0-9;]*m`

**Explanation**:

- `\x1b` - ESC character (ASCII 27, hex 1B)
- `\[` - Literal left bracket
- `[0-9;]*` - Zero or more digits or semicolons (color/style codes)
- `m` - Terminator character

**Common ANSI Codes**:

- `\x1b[0m` - Reset/normal
- `\x1b[1m` - Bold
- `\x1b[31m` - Red foreground
- `\x1b[32m` - Green foreground
- `\x1b[33m` - Yellow foreground
- `\x1b[90m` - Bright black (gray)

## Validation Method

Test implementation in `cmd/gitree/main_test.go` should:

```go
func TestNoColorFlagRemovesANSI(t *testing.T) {
    output := captureOutput(t, "--no-color", testDir)

    // Regex pattern for ANSI escape sequences
    ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
    matches := ansiPattern.FindAllString(output, -1)

    assert.Equal(t, 0, len(matches),
        "Expected zero ANSI sequences, found %d: %v",
        len(matches), matches)
}
```

## Success Criteria

- **SC-002 Validation**: `len(matches) == 0`
- **Scope**: All output to stdout (tree structure, repository paths, git status, messages)
- **Edge Cases**: Errors to stderr should also have no ANSI codes when --no-color is active

## References

- ANSI/VT100 Terminal Control Escape Sequences: <https://www2.ccs.net/~larry/vt100_codes.txt>
- fatih/color NoColor flag: <https://pkg.go.dev/github.com/fatih/color#NoColor>
