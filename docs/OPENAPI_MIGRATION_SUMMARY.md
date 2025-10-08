# OpenAPI Migration Summary

This document summarizes the OpenAPI migration completed for the Go Raspberry Pi Timelapse project.

## What Changed

### 1. New Files Added
- **`openapi.yaml`** - OpenAPI 3.0 specification defining all 14 API endpoints
- **`.spectral.yaml`** - Configuration for OpenAPI linting
- **`codegen-config.yaml`** - Configuration for Go code generation
- **`OPENAPI.md`** - Comprehensive documentation for OpenAPI usage
- **`Makefile`** - Common tasks for development workflow
- **`docs/ADR-001-openapi-migration.md`** - Architecture Decision Record documenting the migration

### 2. Updated Files
- **`.github/workflows/go.yml`** - Added OpenAPI linting and code generation steps
- **`.gitignore`** - Excluded generated code directories
- **`frontend/.gitignore`** - Excluded generated TypeScript code
- **`frontend/package.json`** - Added `generate:api` script and new dependencies
- **`go.mod` & `go.sum`** - Added required Go dependencies for generated code

### 3. Generated Code (Excluded from Git)
- **`generated/server.gen.go`** - Auto-generated Go types from OpenAPI spec
- **`frontend/src/generated/api.ts`** - Auto-generated TypeScript types from OpenAPI spec

## API Endpoints Documented

The OpenAPI specification documents all existing endpoints:

1. **GET `/version`** - Get application version
2. **GET `/capture`** - Capture preview photo
3. **GET `/configuration`** - Get current configuration
4. **POST `/configuration`** - Update configuration
5. **OPTIONS `/configuration`** - CORS preflight for configuration
6. **GET `/monitoring`** - Get system monitoring data
7. **GET `/logs`** - Get application logs
8. **GET `/photos`** - Get list of photos with metadata
9. **GET `/file`** - List all files
10. **GET `/file/last`** - Get most recent file
11. **GET `/file/{fileName}`** - Download specific file
12. **GET `/file/delete`** - Delete files
13. **GET `/archive/zip`** - Download ZIP archive
14. **GET `/archive/tar`** - Download TAR archive
15. **GET `/admin/{command}`** - Execute admin commands

## Quick Start

### Prerequisites
Install required tools globally (one-time setup):
```bash
# Install Spectral for OpenAPI linting
npm install -g @stoplight/spectral-cli

# oapi-codegen is automatically downloaded via 'go run' in CI/CD and Makefile
# No manual installation needed
```

### Development Workflow

#### Using Make (Recommended)
```bash
# Show available commands
make help

# Lint OpenAPI spec and generate all code
make generate

# Run all tests
make test

# Build everything
make build

# Clean generated files
make clean
```

#### Manual Commands

**Lint OpenAPI Specification:**
```bash
spectral lint openapi.yaml
```

**Generate Go Code:**
```bash
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config codegen-config.yaml openapi.yaml
```

**Generate TypeScript Code:**
```bash
cd frontend
npm run generate:api
```

## Benefits Achieved

### ✅ Type Safety
- Both Go and TypeScript have compile-time type checking
- Generated types ensure frontend and backend stay in sync

### ✅ Single Source of Truth
- API definition lives in one place (`openapi.yaml`)
- No more manual duplication of types

### ✅ Automated Validation
- OpenAPI spec is linted on every commit
- Invalid specifications are caught early in CI/CD

### ✅ Better Developer Experience
- IDE autocomplete for API operations
- Type hints for request/response structures
- Clear API documentation

### ✅ Consistency
- Generated code prevents type mismatches
- Frontend TypeScript types match backend Go types exactly

## CI/CD Integration

The GitHub Actions workflow now:
1. Lints the OpenAPI specification
2. Generates Go server code
3. Generates TypeScript client code
4. Builds frontend with generated types
5. Runs all tests (Go and TypeScript)
6. Builds the application

This ensures:
- ✅ API spec is always valid
- ✅ Generated code is always up-to-date
- ✅ No commits with outdated types
- ✅ Both builds use consistent API definitions

## Testing

### Test Coverage
- **Go Tests**: 7 test files, all passing
- **TypeScript Tests**: 91 tests in 8 files, all passing
- **Generated Types**: Dedicated test file verifies TypeScript types

### Running Tests
```bash
# Run Go tests
make test-go

# Run TypeScript tests
make test-ts

# Run all tests
make test
```

## Modifying the API

When you need to change an API endpoint:

1. **Update `openapi.yaml`** with the changes
2. **Lint** the specification: `make lint-api`
3. **Regenerate code**: `make generate`
4. **Update handlers** if needed (in `rest/` package)
5. **Update frontend code** to use new types
6. **Test**: `make test`
7. **Commit** changes (generated code auto-excluded)

## Migration Notes

### No Breaking Changes
- Existing API endpoints remain unchanged
- All handlers in `rest/` package still work as before
- Generated types supplement (not replace) existing code

### Future Enhancements
Consider these improvements in future iterations:
- Generate server stubs for compile-time API implementation validation
- Add Swagger UI for interactive API documentation
- Implement request/response validation middleware
- Add contract testing using the OpenAPI spec

## Resources

- **OpenAPI Specification**: `openapi.yaml`
- **Documentation**: `OPENAPI.md`
- **Architecture Decision**: `docs/ADR-001-openapi-migration.md`
- **Makefile Commands**: Run `make help`

## Support

For issues or questions:
1. Check `OPENAPI.md` for detailed usage instructions
2. Review `docs/ADR-001-openapi-migration.md` for design rationale
3. Run `make help` for available commands
4. Check `.github/workflows/go.yml` for CI/CD pipeline
