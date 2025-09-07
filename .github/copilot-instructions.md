# Cloud SDK - Multi-Cloud Go SDK

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

Cloud SDK is a unified multi-cloud SDK for Go, enabling seamless infrastructure management across AWS, GCP, Azure, and more. The main implementation is in Go with a unified API design inspired by Vercel AI SDK's developer experience.

## Working Effectively

### Prerequisites
- Go 1.22+ required (tested with Go 1.24.6)
- All commands should be run from the `/go` directory within the repository

### Bootstrap and Build
Execute these commands in order for a fresh repository setup:

```bash
cd go/
go mod download      # ~3 seconds - downloads dependencies
go mod verify        # ~1 second - verifies module integrity
go vet ./...         # ~60 seconds - basic linting, NEVER CANCEL
go fmt ./...         # ~1 second - code formatting  
go build ./...       # ~5 seconds - builds all packages
go test -race -cover ./...  # ~90 seconds - runs tests with race detection, NEVER CANCEL, set timeout to 3+ minutes
go mod tidy         # ~1 second - cleans up module dependencies
```

**CRITICAL TIMING NOTES:**
- `go vet ./...` takes ~1 second with cached modules. NEVER CANCEL. Set timeout to 2+ minutes.
- `go test -race -cover ./...` takes ~1-90 seconds (cached vs fresh). NEVER CANCEL. Set timeout to 3+ minutes.
- All other commands complete in under 1 second when modules are cached.

### Linting
**IMPORTANT:** golangci-lint has compatibility issues with Go 1.24+. Use these alternatives:

```bash
# Use these working lint commands:
go vet ./...         # Basic Go linting - works reliably, ~1 second
go fmt ./...         # Code formatting - works reliably, instant

# golangci-lint DOES NOT WORK with Go 1.24+:
# The CI uses golangci-lint v1.55.2 which expects Go 1.22
# Current versions fail with "goarch: unsupported version" errors
# For CI compatibility, stick to go vet and go fmt for local development
```

### Running the Application
**Test the example application:**

```bash
cd examples/aws-sdk/
go run main.go      # ~40 seconds - NEVER CANCEL, set timeout to 2+ minutes
```

**Expected output:** The application will run but show AWS credential errors. This is NORMAL and indicates the application is working correctly. You should see:
- "=== Cloud SDK Example ===" 
- AWS credential error messages for compute, storage, and database operations
- "=== Example completed ==="

## Validation Scenarios

**ALWAYS run these validation steps after making changes:**

1. **Build validation:**
   ```bash
   cd go/
   go build ./...     # Must complete without errors
   ```

2. **Test validation:**
   ```bash
   cd go/
   go test -race -cover ./...  # Must pass all tests, ~90 seconds
   ```

3. **Example validation:**
   ```bash
   cd go/examples/aws-sdk/
   go run main.go     # Must run and show expected AWS credential errors
   ```

4. **Code quality validation:**
   ```bash
   cd go/
   go vet ./...       # Must complete without issues
   go fmt ./...       # Must complete without changes
   go mod tidy        # Must not modify go.mod/go.sum files
   ```

## Repository Structure

### Key Directories
- `go/` - Main Go module and all source code
- `go/providers/aws/` - AWS provider implementation (compute, storage, database)
- `go/services/` - Service interface definitions
- `go/examples/aws-sdk/` - Working example application
- `docs/` - Documentation including comprehensive API docs
- `.github/workflows/` - CI pipeline configuration

### Important Files
- `go/go.mod` - Go module definition, dependencies
- `go/cloudsdk.go` - Main SDK client implementation
- `go/.golangci.yml` - Linting configuration (has compatibility issues)
- `.github/workflows/ci.yml` - CI pipeline (expects Go 1.22)
- `docs/README.md` - Complete API documentation and usage examples

### Key Packages
```
github.com/VAIBHAVSING/Cloudsdk/go              # Main SDK client
github.com/VAIBHAVSING/Cloudsdk/go/providers/aws # AWS provider
github.com/VAIBHAVSING/Cloudsdk/go/services     # Service interfaces  
```

## Common Development Tasks

### Adding a New Provider
1. Create directory in `go/providers/[provider-name]/`
2. Implement the `Provider` interface from `cloudsdk.go`
3. Implement service interfaces from `go/services/`
4. Add tests following existing pattern in `providers/aws/`
5. Run full validation suite

### Adding a New Service
1. Define interface in `go/services/[service].go`
2. Implement in existing providers
3. Add to `Provider` interface in `cloudsdk.go`
4. Add comprehensive tests
5. Update example in `examples/aws-sdk/main.go`

### Before Committing
Always run the complete validation sequence:
```bash
cd go/
go mod tidy && go vet ./... && go fmt ./... && go build ./... && go test -race -cover ./...
```

This ensures CI pipeline compatibility and code quality.

## Troubleshooting

### Build Issues
- **"go.mod not found"**: Ensure you're in the `go/` directory
- **"package not found"**: Run `go mod download` then `go mod tidy`
- **Version conflicts**: Run `go mod tidy` to resolve

### Test Issues  
- **Slow tests**: Tests with `-race` flag take ~90 seconds, this is normal
- **AWS credential errors in examples**: Expected behavior in development environment

### Linting Issues
- **golangci-lint errors**: Use `go vet ./...` and `go fmt ./...` instead
- **Version compatibility**: Project works with Go 1.22+ despite CI using 1.22

## CI Pipeline Information
- Runs on Ubuntu with Go 1.22
- Uses golangci-lint v1.55.2 (may have compatibility issues locally)
- Executes: download, verify, lint, build, test, coverage, tidy check
- All steps must pass for successful builds

## Performance Expectations
- **Total bootstrap time**: ~2 minutes (including tests, first run)
- **Quick development cycle**: ~1 second (build only, cached)
- **Full validation cycle**: ~2 minutes (all lints, tests, builds, cached)
- **Example execution**: ~40 seconds (with AWS timeout retries)

Remember: The SDK is designed for a Vercel AI SDK-like developer experience with clean, simple APIs and easy provider switching.