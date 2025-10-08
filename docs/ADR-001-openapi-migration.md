# ADR-001: Migration to OpenAPI Specification for API Definition and Code Generation

## Status
Accepted

## Context
The Go Raspberry Pi Timelapse application has a REST API that was previously defined through hardcoded routes in `main.go` and handler functions in the `rest` package. The API endpoints, request/response models, and data structures were manually maintained across both the Go backend and TypeScript frontend, leading to potential inconsistencies and duplication of effort.

Key issues with the previous approach:
- API endpoints were hardcoded in `main.go` without formal specification
- Request/response models were manually duplicated in Go (`rest/model.go`) and TypeScript (`frontend/src/models/response.tsx`)
- No automated validation of API contract between frontend and backend
- Risk of type mismatches between client and server
- Manual updates required in multiple places when API changes

## Decision
We have migrated to an OpenAPI 3.0 specification-first approach with automated code generation for both Go server models and TypeScript client code.

### Components

#### 1. OpenAPI Specification (`openapi.yaml`)
- Single source of truth for API definition
- Documents all 14 API endpoints:
  - GET `/version` - Application version
  - GET `/capture` - Capture preview photo
  - GET/POST/OPTIONS `/configuration` - Configuration management
  - GET `/monitoring` - System monitoring data
  - GET `/logs` - Application logs
  - GET `/photos` - List of photos with metadata
  - GET `/file` - List all files
  - GET `/file/last` - Most recent file
  - GET `/file/{fileName}` - Specific file download
  - GET `/file/delete` - Delete files
  - GET `/archive/zip` - ZIP archive download
  - GET `/archive/tar` - TAR archive download
  - GET `/admin/{command}` - Admin commands
- Defines all request/response schemas matching existing Go models

#### 2. OpenAPI Linting
- **Tool**: Spectral CLI
- **Configuration**: `.spectral.yaml`
- **Purpose**: Ensures OpenAPI specification follows best practices and is valid
- Integrated into CI/CD pipeline to validate spec on every commit

#### 3. Go Server Code Generation
- **Tool**: oapi-codegen v2
- **Configuration**: `codegen-config.yaml`
- **Output**: `generated/server.gen.go`
- **Generated artifacts**:
  - Type-safe Go structs for all request/response models
  - Matches existing models in `rest/model.go` and `conf/model.go`
  - Embedded OpenAPI spec for documentation

#### 4. TypeScript Client Code Generation
- **Tool**: openapi-typescript v7
- **Additional dependency**: openapi-fetch for runtime client
- **Output**: `frontend/src/generated/api.ts`
- **Generated artifacts**:
  - TypeScript type definitions for all API operations
  - Type-safe request/response interfaces
  - Path and operation types for compile-time safety

#### 5. CI/CD Integration
Updated `.github/workflows/go.yml` to:
1. Install Spectral CLI and lint OpenAPI spec
2. Install oapi-codegen and generate Go code
3. Generate TypeScript client types
4. Build and test both frontend and backend

The workflow ensures:
- OpenAPI spec is validated before any builds
- Generated code is always in sync with specification
- Both TypeScript and Go code are built after code generation

### Tooling Rationale

**Spectral CLI** - Chosen for OpenAPI linting because:
- Industry-standard linter for OpenAPI/AsyncAPI specs
- Extensible rules engine
- Good CI/CD integration
- Active community support

**oapi-codegen** - Chosen for Go code generation because:
- Native Go implementation
- Excellent support for OpenAPI 3.0
- Generates clean, idiomatic Go code
- No runtime dependencies for generated code
- Flexible configuration options

**openapi-typescript** - Chosen for TypeScript generation because:
- Generates pure TypeScript types (no runtime overhead)
- Works well with modern TypeScript
- Compatible with various fetch clients
- Paired with openapi-fetch for type-safe runtime client

## Consequences

### Positive
- **Single source of truth**: API definition lives in one place (`openapi.yaml`)
- **Type safety**: Compile-time type checking in both Go and TypeScript
- **Consistency**: Generated code ensures frontend and backend always match
- **Documentation**: OpenAPI spec serves as living API documentation
- **Validation**: Automated linting catches specification issues early
- **Reduced errors**: Eliminates manual duplication of types
- **Better DX**: IDE autocomplete and type hints for API operations
- **Future-ready**: Easy to add API documentation UI (Swagger/ReDoc) if needed

### Negative
- **Build complexity**: Additional build steps for code generation
- **Learning curve**: Team needs to understand OpenAPI specification format
- **Generated code**: Need to ignore/exclude generated files from version control
- **Migration effort**: One-time cost to create initial OpenAPI spec

### Neutral
- **Maintenance**: OpenAPI spec must be updated when API changes (replaces manual updates in multiple places)
- **CI/CD time**: Slightly longer build times due to linting and code generation steps

## Implementation Notes
- Generated Go code is placed in `generated/` directory (gitignored)
- Generated TypeScript code is placed in `frontend/src/generated/` directory (gitignored)
- Code generation is integrated into CI/CD before build steps
- Frontend npm scripts include `generate:api` for local development
- Existing API handlers remain unchanged; generated models provide type safety

## Future Improvements
- Consider using generated server interface for compile-time enforcement of API implementation
- Add Swagger UI or ReDoc for interactive API documentation
- Explore contract testing using OpenAPI spec
- Consider adding request/response validation middleware using the spec
