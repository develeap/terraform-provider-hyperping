# Interactive Package

The `interactive` package provides utilities for creating interactive command-line interfaces with guided workflows, real-time feedback, and user-friendly prompts.

## Features

- **Terminal Detection** - Automatically detects TTY environments
- **User Prompts** - String, password, confirmation, and selection prompts
- **Input Validation** - Built-in validators for API keys, file paths, etc.
- **Progress Indicators** - Spinners and progress bars for long operations
- **Graceful Degradation** - Works in non-TTY environments

## Usage

### Basic Prompts

```go
import "github.com/develeap/terraform-provider-hyperping/pkg/interactive"

// Create prompter
prompter := interactive.NewPrompter(interactive.DefaultConfig())

// String input
name, err := prompter.AskString(
    "Enter your name:",
    "default-name",
    "Your full name",
    nil, // optional validator
)

// Password input (hidden)
password, err := prompter.AskPassword(
    "Enter API key:",
    "Get it from: https://example.com/api",
    interactive.APIKeyValidator,
)

// Confirmation
proceed, err := prompter.AskConfirm("Continue?", true)

// Selection
choice, err := prompter.AskSelect(
    "Choose mode:",
    []string{"Full", "Dry-run", "Validate"},
    "Full",
)
```

### Input Validation

```go
// API key validation
err := interactive.APIKeyValidator("test-key")

// Hyperping key validation (must start with sk_)
err := interactive.HyperpingAPIKeyValidator("sk_test123")

// Platform-specific validation
validator := interactive.SourceAPIKeyValidator("betterstack")
err := validator("my-api-key")

// File path validation
err := interactive.FilePathValidator("/path/to/file")
```

### Progress Indicators

```go
// Spinner for indeterminate operations
spinner := interactive.NewSpinner("Loading...", os.Stderr)
spinner.Start()
// ... do work ...
spinner.SuccessMessage("Done!")

// Progress bar for batch operations
bar := interactive.NewProgressBar(100, "Processing", os.Stderr)
for i := 0; i < 100; i++ {
    // ... do work ...
    bar.Add(1)
}
bar.Finish()
```

### Terminal Detection

```go
// Check if running in interactive terminal
if interactive.IsInteractive() {
    // Use interactive mode
} else {
    // Use non-interactive mode
}

// Check for ANSI support
if interactive.SupportsANSI() {
    // Use colors
}
```

## API Reference

### Prompter

```go
type Prompter struct { ... }

func NewPrompter(config *Config) *Prompter
func (p *Prompter) AskString(message, defaultValue, help string, validator survey.Validator) (string, error)
func (p *Prompter) AskPassword(message, help string, validator survey.Validator) (string, error)
func (p *Prompter) AskConfirm(message string, defaultValue bool) (bool, error)
func (p *Prompter) AskSelect(message string, options []string, defaultValue string) (string, error)
func (p *Prompter) PrintHeader(title string)
func (p *Prompter) PrintSuccess(message string)
func (p *Prompter) PrintError(message string)
func (p *Prompter) PrintWarning(message string)
func (p *Prompter) PrintInfo(message string)
```

### Validators

```go
func APIKeyValidator(val interface{}) error
func HyperpingAPIKeyValidator(val interface{}) error
func SourceAPIKeyValidator(platform string) survey.Validator
func FilePathValidator(val interface{}) error
```

### Progress Indicators

```go
type Spinner struct { ... }
func NewSpinner(message string, writer io.Writer) *Spinner
func (s *Spinner) Start()
func (s *Spinner) Stop()
func (s *Spinner) UpdateMessage(message string)
func (s *Spinner) SuccessMessage(message string)
func (s *Spinner) ErrorMessage(message string)

type ProgressBar struct { ... }
func NewProgressBar(maxValue int64, description string, writer io.Writer) *ProgressBar
func (pb *ProgressBar) Add(n int) error
func (pb *ProgressBar) Set(n int) error
func (pb *ProgressBar) Finish() error
```

### Terminal Detection

```go
func IsInteractive() bool
func IsTTY(w io.Writer) bool
func SupportsANSI() bool
```

## Examples

See the migration tools for complete examples:

- `/cmd/migrate-betterstack/interactive.go`
- `/cmd/migrate-uptimerobot/interactive.go`
- `/cmd/migrate-pingdom/interactive.go`

## Testing

Run tests:

```bash
go test ./pkg/interactive/... -v
```

Run with coverage:

```bash
go test ./pkg/interactive/... -cover
```

## Dependencies

- `github.com/AlecAivazis/survey/v2` - User prompts
- `github.com/briandowns/spinner` - Loading spinners
- `github.com/schollz/progressbar/v3` - Progress bars
- `github.com/mattn/go-isatty` - Terminal detection

## Best Practices

1. **Always provide defaults** for string inputs
2. **Use validators** to catch errors early
3. **Show help text** for complex inputs
4. **Test in both TTY and non-TTY** environments
5. **Handle Ctrl+C gracefully** - all prompts return error on interrupt
6. **Use appropriate indicators** - spinners for unknown duration, progress bars for known

## Error Handling

All prompt methods return `errors.New("operation cancelled by user")` when the user presses Ctrl+C. Handle this gracefully:

```go
result, err := prompter.AskString("Enter value:", "", "", nil)
if err != nil {
    if err.Error() == "operation cancelled by user" {
        fmt.Println("Operation cancelled")
        return 0
    }
    return 1
}
```

## Environment Variables

The package respects standard environment variables:

- `NO_COLOR` - Disables ANSI colors when set

## License

Copyright (c) 2026 Develeap
SPDX-License-Identifier: MPL-2.0
