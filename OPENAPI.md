# OpenAPI Code Generation

This project uses OpenAPI 3.0 specification to define the API and generate both Go server models and TypeScript client code.

## Overview

- **OpenAPI Spec**: `openapi.yaml` - Single source of truth for API definition
- **Go Generated Code**: `generated/server.gen.go` - Auto-generated from OpenAPI spec
- **TypeScript Generated Code**: `frontend/src/generated/api.ts` - Auto-generated from OpenAPI spec

## Prerequisites

### For Go Code Generation
```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

### For TypeScript Code Generation
The required packages are already in `frontend/package.json`:
- `openapi-typescript` (dev dependency)
- `openapi-fetch` (runtime dependency)

### For OpenAPI Linting
```bash
npm install -g @stoplight/spectral-cli
```

## Usage

### Linting the OpenAPI Specification
```bash
spectral lint openapi.yaml
```

### Generating Go Server Code
```bash
oapi-codegen -config codegen-config.yaml openapi.yaml
```

This generates type-safe Go structs in `generated/server.gen.go` matching the OpenAPI schemas.

### Generating TypeScript Client Code
```bash
cd frontend
npm run generate:api
```

This generates TypeScript type definitions in `frontend/src/generated/api.ts`.

### Full Build Process (as in CI/CD)
```bash
# 1. Lint OpenAPI spec
spectral lint openapi.yaml

# 2. Generate Go code
oapi-codegen -config codegen-config.yaml openapi.yaml

# 3. Generate TypeScript code
cd frontend && npm run generate:api && cd ..

# 4. Build frontend
cd frontend && npm run build && cd ..

# 5. Build Go application
go build -v .
```

## Working with the Generated Code

### Go
The generated Go code provides type-safe structs for all request/response models:

```go
import "github.com/ingojaeckel/go-raspberry-pi-timelapse/generated"

// Use generated types
var settings generated.Settings
var monitoring generated.MonitoringResponse
```

### TypeScript
The generated TypeScript types can be imported and used with type safety:

```typescript
import { components } from './generated/api';

type Settings = components['schemas']['Settings'];
type MonitoringResponse = components['schemas']['MonitoringResponse'];
```

For making API calls with full type safety, you can use `openapi-fetch`:

```typescript
import createClient from 'openapi-fetch';
import type { paths } from './generated/api';

const client = createClient<paths>({ baseUrl: 'http://localhost:8080' });

// Type-safe API calls
const { data, error } = await client.GET('/configuration');
const { data, error } = await client.POST('/configuration', {
  body: { /* Settings object */ }
});
```

## Modifying the API

When you need to change the API:

1. **Update `openapi.yaml`** - Make changes to the OpenAPI specification
2. **Lint the spec** - Ensure it's valid: `spectral lint openapi.yaml`
3. **Regenerate code** - Run code generation for both Go and TypeScript
4. **Update handlers** - If needed, update the Go handler functions in `rest/` package
5. **Test** - Run both Go and TypeScript tests to ensure compatibility

## CI/CD Integration

The GitHub Actions workflow (`.github/workflows/go.yml`) automatically:
1. Lints the OpenAPI specification
2. Generates Go and TypeScript code
3. Builds and tests both frontend and backend

This ensures the API specification and generated code are always in sync.

## Benefits

- **Type Safety**: Both Go and TypeScript have compile-time type checking
- **Single Source of Truth**: API definition lives in one place
- **Consistency**: Generated code ensures frontend and backend always match
- **Documentation**: OpenAPI spec serves as living API documentation
- **Validation**: Automated linting catches specification issues early

## Troubleshooting

### "no required module provides package github.com/getkin/kin-openapi/openapi3"
Run: `go get github.com/getkin/kin-openapi/openapi3`

### TypeScript generation fails
Ensure you're in the frontend directory and dependencies are installed:
```bash
cd frontend
npm install --legacy-peer-deps
npm run generate:api
```

### OpenAPI lint errors
Check the `.spectral.yaml` configuration and ensure your OpenAPI spec follows the rules.
