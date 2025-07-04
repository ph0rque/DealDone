## Relevant Files

- `cmd/dealdone/main.go` - New application entry point replacing current main.go
- `internal/app/app.go` - Refactored App struct and core application logic
- `internal/app/config.go` - Centralized configuration management
- `internal/app/startup.go` - Service initialization and dependency injection
- `internal/core/templates/types.go` - Template-specific type definitions
- `internal/core/templates/service.go` - Unified template service interface
- `internal/core/documents/types.go` - Document-specific type definitions
- `internal/core/documents/service.go` - Document service interface
- `internal/infrastructure/webhooks/types.go` - Webhook type definitions
- `internal/infrastructure/ai/types.go` - AI-specific type definitions
- `internal/shared/types/errors.go` - Common error types
- `internal/shared/types/results.go` - Common result types
- `internal/shared/types/metadata.go` - Common metadata types
- `internal/shared/logger/logger.go` - Logger interface
- `internal/shared/utils/utils.go` - Common utility functions
- `go.mod` - Updated module configuration

### Notes

- This refactoring requires careful coordination to maintain functionality during migration
- Each package should have its own test files alongside the implementation
- Use feature flags or parallel structure to enable gradual migration
- Ensure all existing tests continue to pass throughout the refactoring

## Parent Tasks

- [x] 1.0 Set up new package structure and foundation ✅
- [x] 2.0 Refactor core business logic packages ✅
- [x] 3.0 Refactor infrastructure packages ✅
- [ ] 4.0 Refactor application layer
- [ ] 5.0 Complete migration and cleanup

---

## 1.0 Set up new package structure and foundation ✅

### 1.1 Create directory structure ✅
- Create `internal/` directory with subdirectories:
  - `app/` - Application layer
  - `core/` - Core business logic
  - `infrastructure/` - External integrations
  - `domain/` - Domain models and business rules
  - `shared/` - Shared utilities and types

### 1.2 Create `cmd/dealdone/main.go` ✅
- Move minimal main function from current `main.go`
- Keep only Wails initialization
- Delegate all logic to `internal/app`

### 1.3 Create package documentation files ✅
- Add `doc.go` to each new package
- Document package purpose and responsibilities

### 1.4 Set up shared types package ✅
- Create `internal/shared/types/errors.go` for common error types
- Create `internal/shared/types/results.go` for common result types
- Create `internal/shared/types/metadata.go` for common metadata

### 1.5 Create logger interface ✅
- Define logger interface in `internal/shared/logger`
- Will be implemented by App layer

### 1.6 Verify build still works ✅
- Ensure project compiles with new structure
- No functionality changes yet

### 1.7 Update imports in existing files ✅
- Update any files that reference moved types
- Ensure all imports are correct

---

## 2.0 Refactor core business logic packages ✅

### 2.1 Move template-related files ✅
- Move all template*.go files to `internal/core/templates/`
- Update package declarations to `package templates`

### 2.2 Extract template types ✅
- Create `internal/core/templates/types.go`
- Move template-specific types from `types.go`
- Add any missing template-related types

### 2.3 Create template service interface ✅
- Define `Service` interface in `internal/core/templates/service.go`
- Combine functionality from all template files
- Create unified API for template operations

### 2.4 Move document processing files ✅
- Move `documentprocessor.go` and `documentrouter.go` to `internal/core/documents/`
- Update package declarations

### 2.5 Extract document types ✅
- Create `internal/core/documents/types.go`
- Move document-specific types from `types.go`

### 2.6 Create document service interface ✅
- Define `Service` interface in `internal/core/documents/service.go`
- Unify document processing and routing

### 2.7 Move deal/folder management ✅
- Move `foldermanager.go` to `internal/core/deals/`
- Create deal management service

### 2.8 Extract deal types ✅
- Create `internal/core/deals/types.go`
- Define Deal, DealFolder, and related types

### 2.9 Update cross-package references ✅
- Fix any broken imports between core packages
- Ensure proper dependency direction

### 2.10 Verify core packages compile ✅
- Test that all core packages build correctly
- No integration yet, just compilation

---

## 3.0 Refactor infrastructure packages ✅

### 3.1 Move AI provider files ✅
- Move `aiprovider_*.go`, `aiservice.go` to `internal/infrastructure/ai/`
- Move `aicache.go` to same package

### 3.2 Extract AI types ✅
- Create `internal/infrastructure/ai/types.go`
- Move AI-specific types from `types.go`

### 3.3 Create AI service interface ✅
- Define unified AI service interface
- Abstract provider implementations

### 3.4 Move webhook files ✅
- Move `webhook*.go` files to `internal/infrastructure/webhooks/`
- Keep webhook schemas together

### 3.5 Extract webhook types ✅
- Create `internal/infrastructure/webhooks/types.go`
- Move webhook-specific types from `types.go`

### 3.6 Move n8n integration ✅
- Move `n8nintegration.go` to `internal/infrastructure/n8n/`
- Create n8n service interface

### 3.7 Move storage/file operations ✅
- Move `filemanager.go` to `internal/infrastructure/storage/`
- Create storage abstraction layer

### 3.8 Move OCR service ✅
- Move `ocrservice.go` to `internal/infrastructure/ocr/`
- Define OCR service interface

### 3.9 Update infrastructure dependencies ✅
- Fix imports between infrastructure packages
- Ensure no circular dependencies

### 3.10 Create infrastructure factories ✅
- Create factory functions for each infrastructure service
- Prepare for dependency injection

### 3.11 Verify infrastructure packages compile ✅
- Test that all infrastructure packages build
- Check interface implementations

---

## 4.0 Refactor application layer

### 4.1 Move App struct
- Create `internal/app/app.go`
- Move App struct and core methods
- Keep only orchestration logic

### 4.2 Create configuration management
- Create `internal/app/config.go`
- Move all configuration logic
- Centralize config handling

### 4.3 Implement dependency injection
- Create `internal/app/startup.go`
- Initialize all services with proper dependencies
- Wire up service interfaces

### 4.4 Move domain analysis files
- Move `dealvaluation.go`, `competitiveanalysis.go`, etc. to `internal/domain/analysis/`
- Create analysis service

### 4.5 Move queue management
- Move `queuemanager.go` to `internal/domain/queue/`
- Move `jobtracker.go` to same location

### 4.6 Move workflow management
- Move `workflowrecovery.go` to `internal/domain/workflow/`
- Create workflow service

### 4.7 Create domain service interfaces
- Define interfaces for each domain service
- Ensure clean separation from infrastructure

### 4.8 Update App to use interfaces
- Replace direct struct usage with interfaces
- Implement proper service delegation

### 4.9 Move utility files
- Move `utils.go` to `internal/shared/utils/`
- Move any other utility functions

### 4.10 Verify application layer
- Ensure App properly orchestrates all services
- Test initialization and startup

---

## 5.0 Complete migration and cleanup

### 5.1 Delete old types.go
- Ensure all types have been moved
- Remove the monolithic types.go file

### 5.2 Update tests
- Move test files to appropriate packages
- Update test imports
- Ensure all tests pass

### 5.3 Update build configuration
- Update `wails.json` if needed
- Update any build scripts

### 5.4 Create package dependency diagram
- Document the new architecture
- Show clean dependency flow

### 5.5 Update developer documentation
- Update README with new structure
- Document how to add new features
- Explain package responsibilities

### 5.6 Performance testing
- Ensure no performance regression
- Test startup time
- Verify memory usage

### 5.7 Integration testing
- Full end-to-end testing
- Verify all workflows still function
- Test with real data

### 5.8 Update CI/CD
- Update any CI/CD pipelines
- Ensure automated tests pass

### 5.9 Code review preparation
- Create summary of changes
- Document migration decisions
- Prepare for team review

### 5.10 Final cleanup
- Remove any temporary files
- Clean up any migration artifacts
- Final build and test