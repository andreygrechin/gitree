# Contracts

This feature does not introduce external API contracts. The colorization changes are internal to the `GitStatus.Format()` method.

## Output Format Contract

While not an external API, the output format serves as an informal contract with users who may parse the output.

### Format Change

**Before (spec 001)**:
```
[branch status-indicators]
```

**After (spec 002)**:
```
[[ branch | status-indicators ]]
```

### ANSI Color Codes

When terminal supports color (TTY + NO_COLOR not set):
- Output includes ANSI escape sequences (`\033[XXm`)
- Color codes automatically stripped for non-TTY output (pipes, redirects)

### Examples

**With colors (TTY)**:
```
\033[90m[[\033[0m \033[90mmain\033[0m \033[90m]]\033[0m
\033[90m[[\033[0m \033[33mfeature/test\033[0m \033[90m|\033[0m \033[32m↑2\033[0m \033[90m]]\033[0m
```

**Without colors (pipe/redirect or NO_COLOR)**:
```
[[ main ]]
[[ feature/test | ↑2 ]]
```

## Internal Contracts

### GitStatus.Format() Method Signature

**Unchanged**:
```go
func (g *GitStatus) Format() string
```

**Behavior Change**:
- Returns colorized output with double brackets
- Respects terminal capability detection
- Honors NO_COLOR environment variable

## Dependencies

### New Explicit Dependency
- `github.com/fatih/color@v1.18.0` (upgraded from transitive v1.7.0)

### Environment Variables
- `NO_COLOR`: When set (any non-empty value), colors are disabled
- Standard: https://no-color.org/
