# Documentation Index

This directory contains comprehensive documentation for the Go Raspberry Pi Timelapse project.

## Quick Navigation

### OpenAPI Integration
The project uses OpenAPI 3.0 for API specification and code generation.

📘 **[OPENAPI_ARCHITECTURE.md](OPENAPI_ARCHITECTURE.md)** - Visual architecture overview  
Start here for a high-level understanding of the OpenAPI integration with diagrams.

📘 **[OPENAPI_MIGRATION_SUMMARY.md](OPENAPI_MIGRATION_SUMMARY.md)** - Migration summary  
Complete guide to what changed and how to use the new OpenAPI-based workflow.

📘 **[ADR-001-openapi-migration.md](ADR-001-openapi-migration.md)** - Architecture Decision Record  
Detailed rationale for the OpenAPI migration, including tooling choices and trade-offs.

### Project Images
- **build.JPG** - Overview of the physical build
- **build_detail.JPG** - Detailed view of the components
- **go-raspberry-pi-timelapse.jpg** - Project image

## Related Documentation

### In Root Directory
- **[../OPENAPI.md](../OPENAPI.md)** - Usage guide for OpenAPI features
- **[../README.md](../README.md)** - Main project README

## For New Developers

If you're new to the project and want to understand the API architecture:

1. Read **OPENAPI_ARCHITECTURE.md** for the big picture
2. Review **OPENAPI_MIGRATION_SUMMARY.md** for practical usage
3. Check **ADR-001-openapi-migration.md** for design decisions
4. Run `./validate-openapi.sh` to verify your setup

## For API Changes

When modifying the API:

1. Update `../openapi.yaml`
2. Run `make generate` to regenerate code
3. Update handlers in `../rest/` if needed
4. Run tests with `make test`
5. Update documentation if needed

See [../OPENAPI.md](../OPENAPI.md) for detailed instructions.
