# Go Coding Standards

## Table of Contents
- [General Principles](#general-principles)
- [Code Formatting](#code-formatting)
- [Naming Conventions](#naming-conventions)
- [Package Organization](#package-organization)
- [Error Handling](#error-handling)
- [Comments and Documentation](#comments-and-documentation)
- [Testing](#testing)
- [Performance and Best Practices](#performance-and-best-practices)
- [Security](#security)
- [Tools and Automation](#tools-and-automation)

## General Principles

### Embrace Go Idioms
- Write idiomatic Go code that follows community conventions
- Favor composition over inheritance
- Use interfaces to define behavior, not data
- Keep it simple and readable over clever

### Code Should Be Self-Documenting
- Choose clear, descriptive names
- Write code that explains itself
- Add comments only when the "why" isn't obvious

## Code Formatting

### Use Standard Tools
```bash
# Always format code before committing
go fmt ./...

# Use goimports for import management
goimports -w .

# Run the linter
golangci-lint run
```

### Line Length and Wrapping
- Aim for 80-100 characters per line
- Break long function signatures sensibly:
```go
// Good
func ProcessUserData(
    ctx context.Context,
    userID string,
    options *ProcessingOptions,
) (*UserResult, error) {
    // implementation
}
```

### Imports Organization
```go
// Standard library first
import (
    "context"
    "fmt"
    "time"
    
    // Third-party packages
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    
    // Local packages
    "myproject/internal/config"
    "myproject/pkg/utils"
)
```

## Naming Conventions

### Variables and Functions
- Use camelCase for unexported names: `userService`, `processData`
- Use PascalCase for exported names: `UserService`, `ProcessData`
- Use short, clear names in limited scope: `i` for loop counters, `ctx` for context
- Use descriptive names for broader scope: `userRepository`, `configurationManager`

### Constants
```go
// Use camelCase for unexported constants
const defaultTimeout = 30 * time.Second

// Use PascalCase for exported constants
const MaxRetryAttempts = 3

// Use iota for enumerations
type Status int

const (
    StatusPending Status = iota
    StatusProcessing
    StatusCompleted
    StatusFailed
)
```

### Packages
- Use short, clear, lowercase names
- Avoid stuttering: `user.User` not `user.UserStruct`
- Package names should be singular: `user`, not `users`

## Package Organization

### Project Structure
```
myproject/
├── cmd/                    # Main applications
│   └── server/
│       └── main.go
├── internal/               # Private application code
│   ├── config/
│   ├── handler/
│   ├── service/
│   └── repository/
├── pkg/                    # Public library code
│   └── utils/
├── api/                    # API definitions (OpenAPI, protobuf)
├── web/                    # Web app assets
├── scripts/               # Build and deployment scripts
├── test/                  # Test data and utilities
├── go.mod
├── go.sum
└── README.md
```

### Package Guidelines
- Keep packages focused on a single responsibility
- Minimize dependencies between packages
- Use `internal/` for code that shouldn't be imported by other projects
- Use `pkg/` for reusable library code

## Error Handling

### Error Creation and Wrapping
```go
// Create errors with context
if err != nil {
    return fmt.Errorf("failed to process user %s: %w", userID, err)
}

// Use custom error types for specific cases
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

### Error Checking
```go
// Check errors immediately
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Don't ignore errors
_ = file.Close() // Bad
if err := file.Close(); err != nil {
    log.Printf("failed to close file: %v", err)
}
```

## Comments and Documentation

### Package Documentation
```go
// Package user provides functionality for user management and authentication.
// It includes user creation, validation, and session management capabilities.
package user
```

### Function Documentation
```go
// ProcessUserData validates and processes user information.
// It returns a UserResult containing processed data and any validation errors.
//
// The context should include a timeout for external API calls.
// Options can be nil for default processing behavior.
func ProcessUserData(ctx context.Context, data *UserData, options *ProcessingOptions) (*UserResult, error) {
    // implementation
}
```

### Inline Comments
```go
// Only comment the "why", not the "what"
if user.Age < 18 {
    // Legal compliance: minors require parental consent
    return requireParentalConsent(user)
}
```

## Testing

### Test File Organization
```go
// user_test.go
package user

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

### Test Naming and Structure
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateUserRequest
        want    *User
        wantErr bool
    }{
        {
            name: "valid user creation",
            input: &CreateUserRequest{
                Email: "test@example.com",
                Name:  "Test User",
            },
            want: &User{
                Email: "test@example.com",
                Name:  "Test User",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            input: &CreateUserRequest{
                Email: "invalid-email",
                Name:  "Test User",
            },
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewUserService()
            got, err := service.CreateUser(tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Test Coverage
- Aim for 80%+ coverage on business logic
- Test happy paths and error cases
- Use table-driven tests for multiple scenarios
- Mock external dependencies

## Performance and Best Practices

### Memory Management
```go
// Reuse slices when possible
func processItems(items []Item) []ProcessedItem {
    // Pre-allocate with known capacity
    results := make([]ProcessedItem, 0, len(items))
    
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// Use string.Builder for string concatenation
func buildMessage(parts []string) string {
    var builder strings.Builder
    builder.Grow(estimatedSize) // Pre-allocate if size is known
    
    for _, part := range parts {
        builder.WriteString(part)
    }
    return builder.String()
}
```

### Goroutines and Concurrency
```go
// Use context for cancellation
func processWithTimeout(ctx context.Context, data []Item) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Use worker pool pattern for heavy processing
    jobs := make(chan Item, len(data))
    results := make(chan Result, len(data))
    
    // Start workers
    for i := 0; i < runtime.NumCPU(); i++ {
        go worker(ctx, jobs, results)
    }
    
    // Send jobs
    go func() {
        defer close(jobs)
        for _, item := range data {
            jobs <- item
        }
    }()
    
    // Collect results
    for i := 0; i < len(data); i++ {
        select {
        case result := <-results:
            // handle result
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return nil
}
```

### Interface Usage
```go
// Define interfaces where they're used, not where they're implemented
type UserStore interface {
    GetUser(id string) (*User, error)
    SaveUser(*User) error
}

// Keep interfaces small and focused
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}
```

## Security

### Input Validation
```go
func validateEmail(email string) error {
    if len(email) == 0 {
        return errors.New("email is required")
    }
    
    if len(email) > 254 {
        return errors.New("email too long")
    }
    
    // Use regex or specialized library for validation
    if !emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }
    
    return nil
}
```

### Secrets Management
```go
// Never hardcode secrets
const (
    // Bad
    apiKey = "sk-1234567890abcdef"
    
    // Good - use environment variables or config
    defaultPort = 8080
)

// Load from environment
func loadConfig() *Config {
    return &Config{
        APIKey:      os.Getenv("API_KEY"),
        DatabaseURL: os.Getenv("DATABASE_URL"),
        Port:        getEnvInt("PORT", defaultPort),
    }
}
```

## Tools and Automation

### Required Tools
```bash
# Install essential tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/sast-scan@latest
```

### Pre-commit Hook Example
```bash
#!/bin/sh
# .git/hooks/pre-commit

set -e

echo "Running Go checks..."

# Format code
goimports -w .

# Vet code
go vet ./...

# Run tests
go test ./...

# Lint code
golangci-lint run

echo "All checks passed!"
```

### Makefile
```makefile
.PHONY: test build lint fmt

fmt:
	goimports -w .
	go fmt ./...

lint:
	golangci-lint run

test:
	go test -race -cover ./...

build:
	go build -o bin/myapp ./cmd/server

check: fmt lint test

ci: check build
```

---

## Quick Reference Checklist

Before committing code, ensure:
- [ ] Code is formatted with `gofmt` and `goimports`
- [ ] All tests pass
- [ ] Linter shows no issues
- [ ] Error handling is proper and consistent
- [ ] Function and package documentation is complete
- [ ] No hardcoded secrets or credentials
- [ ] Appropriate test coverage for new functionality