# Product Requirements Document: Backend Architecture Refactoring

**Version:** 2.0  
**Date:** January 2025  
**Status:** Planning  
**Dependencies:** None (Refactoring initiative)

## Executive Summary

PRD 2.0 addresses critical technical debt in DealDone's backend architecture by refactoring the monolithic Go package structure into a well-organized, modular system. The current architecture has all backend logic dumped into a single `main` package with a massive `types.go` file containing 766 lines of mixed type definitions. This refactoring will improve code maintainability, testability, and developer experience while enabling future scalability.

## Current State Analysis

### Critical Issues

1. **Monolithic Package Structure**
   - All 50+ Go files in root directory under single `main` package
   - No separation of concerns or logical grouping
   - Difficult to navigate and understand code organization
   - Circular dependency risks and tight coupling

2. **Types.go Anti-pattern**
   - Single 766-line file containing ALL type definitions
   - Types mixed from different domains (webhooks, templates, files, AI, etc.)
   - No co-location of types with their related logic
   - Shared types not properly elevated to common packages

3. **Template-Related Code Scattered**
   - 11 template-related files mixed with other concerns
   - No cohesive template package despite clear domain boundary
   - Template types spread across multiple files

4. **Service Initialization Chaos**
   - 30+ services initialized in app.go startup method
   - Complex interdependencies not clearly expressed
   - Difficult to test services in isolation

## Goals

### Primary Objectives
1. **Establish Clear Package Structure** with domain-driven organization
2. **Eliminate types.go** by distributing types to appropriate packages
3. **Create Focused Packages** for major domains (templates, webhooks, documents, etc.)
4. **Implement Layered Architecture** with clear boundaries and dependencies
5. **Improve Testability** through proper dependency injection and interfaces

### Success Criteria
- Zero files in root directory except main.go and configuration
- All types co-located with their related logic
- Clear package dependency graph with no circular dependencies
- Improved code navigation and discoverability
- Reduced coupling between components
- Enhanced unit test coverage due to better isolation

## Proposed Architecture

### Package Structure
```
DealDone/
├── cmd/
│   └── dealdone/
│       └── main.go              # Application entry point
├── internal/                    # Private application code
│   ├── app/                     # Application layer
│   │   ├── app.go              # App struct and initialization
│   │   ├── config.go           # Configuration management
│   │   └── startup.go          # Service initialization
│   ├── core/                   # Core business logic
│   │   ├── deals/              # Deal management
│   │   │   ├── deal.go         # Deal types and logic
│   │   │   ├── folder.go       # Folder management
│   │   │   └── service.go      # Deal service
│   │   ├── documents/          # Document processing
│   │   │   ├── document.go     # Document types
│   │   │   ├── processor.go    # Document processor
│   │   │   ├── router.go       # Document router
│   │   │   └── service.go      # Document service
│   │   └── templates/          # Template management
│   │       ├── types.go        # Template-specific types
│   │       ├── manager.go      # Template manager
│   │       ├── parser.go       # Template parser
│   │       ├── populator.go    # Template populator
│   │       ├── discovery.go    # Template discovery
│   │       ├── optimizer.go    # Template optimizer
│   │       ├── analytics.go    # Template analytics
│   │       └── service.go      # Template service facade
│   ├── infrastructure/         # External integrations
│   │   ├── ai/                 # AI providers
│   │   │   ├── types.go        # AI-specific types
│   │   │   ├── service.go      # AI service interface
│   │   │   ├── openai.go       # OpenAI provider
│   │   │   ├── claude.go       # Claude provider
│   │   │   ├── cache.go        # AI response cache
│   │   │   └── config.go       # AI configuration
│   │   ├── webhooks/           # Webhook system
│   │   │   ├── types.go        # Webhook types
│   │   │   ├── service.go      # Webhook service
│   │   │   ├── handlers.go     # Webhook handlers
│   │   │   ├── schemas.go      # Schema validation
│   │   │   └── auth.go         # Webhook authentication
│   │   ├── n8n/                # n8n integration
│   │   │   ├── types.go        # n8n-specific types
│   │   │   ├── client.go       # n8n API client
│   │   │   └── service.go      # n8n service
│   │   └── storage/            # Storage layer
│   │       ├── filesystem.go   # File system operations
│   │       └── persistence.go  # Data persistence
│   ├── domain/                 # Domain models and logic
│   │   ├── analysis/           # Analysis features
│   │   │   ├── valuation.go    # Deal valuation
│   │   │   ├── competitive.go  # Competitive analysis
│   │   │   ├── trends.go       # Trend analysis
│   │   │   └── anomaly.go      # Anomaly detection
│   │   ├── queue/              # Queue management
│   │   │   ├── types.go        # Queue types
│   │   │   ├── manager.go      # Queue manager
│   │   │   └── persistence.go  # Queue persistence
│   │   └── workflow/           # Workflow management
│   │       ├── types.go        # Workflow types
│   │       ├── recovery.go     # Workflow recovery
│   │       └── executor.go     # Workflow executor
│   └── shared/                 # Shared utilities
│       ├── types/              # Common types
│       │   ├── errors.go       # Error types
│       │   ├── results.go      # Result types
│       │   └── metadata.go     # Metadata types
│       ├── utils/              # Utility functions
│       │   ├── validation.go   # Validation helpers
│       │   ├── formatting.go   # Formatting utilities
│       │   └── crypto.go       # Cryptographic utilities
│       └── logger/             # Logging infrastructure
│           └── logger.go       # Logger interface
├── pkg/                        # Public packages (if needed)
│   └── client/                 # Client SDK
├── frontend/                   # React frontend (unchanged)
├── deployment/                 # Deployment utilities
├── monitoring/                 # Monitoring utilities  
├── performance/                # Performance utilities
├── testing/                    # Test utilities
├── n8n-workflows/             # n8n workflow definitions
├── memory-bank/               # Project documentation
├── tasks/                     # Project tasks and PRDs
├── go.mod
├── go.sum
└── README.md
```

### Type Distribution Strategy

#### 1. Domain-Specific Types
Each package contains its own types that are closely related to its functionality:

**internal/core/templates/types.go:**
```go
package templates

type Template struct {
    ID          string
    Name        string
    Path        string
    Type        TemplateType
    Fields      []TemplateField
    Metadata    TemplateMetadata
}

type TemplateField struct {
    Name        string
    Type        FieldType
    Required    bool
    Validators  []FieldValidator
}

// ... other template-specific types
```

**internal/infrastructure/webhooks/types.go:**
```go
package webhooks

type WebhookPayload struct {
    JobID        string
    DealName     string
    TriggerType  TriggerType
    WorkflowType WorkflowType
    // ... webhook-specific fields
}

type WebhookResult struct {
    Status    string
    Results   ProcessingResults
    Errors    []ProcessingError
}

// ... other webhook-specific types
```

#### 2. Shared Types
Common types used across packages go in shared packages:

**internal/shared/types/results.go:**
```go
package types

type Result[T any] struct {
    Value T
    Error error
}

type ProcessingResult struct {
    Success bool
    Message string
    Data    interface{}
}
```

#### 3. Interface Definitions
Each service defines its interface in the same package:

**internal/core/documents/service.go:**
```go
package documents

type Service interface {
    ProcessDocument(ctx context.Context, path string) (*Document, error)
    RouteDocument(ctx context.Context, doc *Document, dealName string) (*RoutingResult, error)
}
```

### Dependency Rules

1. **Layered Architecture**:
   - `cmd` → `internal/app` → `internal/core` → `internal/infrastructure`
   - `internal/domain` can be used by `core` and `infrastructure`
   - `internal/shared` can be used by all layers

2. **No Circular Dependencies**:
   - Dependencies flow downward only
   - Use interfaces to invert dependencies when needed

3. **Package Cohesion**:
   - Each package has a single, well-defined responsibility
   - High cohesion within packages, loose coupling between packages

## Implementation Plan

### Phase 1: Foundation (Week 1)
1. **Create New Directory Structure**
   - Set up all directories as specified
   - Create package documentation files
   - Update go.mod with new module structure

2. **Move main.go**
   - Create `cmd/dealdone/main.go`
   - Minimal main function that calls app initialization

3. **Extract Shared Types**
   - Create `internal/shared/types` package
   - Move common error types, result types, metadata types
   - Update imports across codebase

### Phase 2: Core Refactoring (Week 2-3)
1. **Refactor Template Package**
   - Move all template-related files to `internal/core/templates`
   - Extract template types from types.go
   - Create unified template service interface
   - Update all template-related imports

2. **Refactor Document Package**
   - Move document processing files to `internal/core/documents`
   - Extract document types from types.go
   - Create document service interface

3. **Refactor Deal Package**
   - Move deal and folder management to `internal/core/deals`
   - Extract deal-related types
   - Create deal service interface

### Phase 3: Infrastructure (Week 3-4)
1. **Refactor AI Package**
   - Move AI providers to `internal/infrastructure/ai`
   - Extract AI types from types.go
   - Implement provider pattern with clear interfaces

2. **Refactor Webhook Package**
   - Move webhook system to `internal/infrastructure/webhooks`
   - Extract webhook types (largest section in types.go)
   - Separate concerns: handlers, validation, authentication

3. **Refactor n8n Integration**
   - Move to `internal/infrastructure/n8n`
   - Extract n8n-specific types
   - Clean separation from webhook system

### Phase 4: App Layer (Week 4-5)
1. **Refactor App Structure**
   - Move app.go to `internal/app`
   - Extract service initialization to `startup.go`
   - Implement dependency injection pattern

2. **Configuration Management**
   - Centralize configuration in `internal/app/config.go`
   - Environment-specific configurations
   - Validation and defaults

### Phase 5: Testing and Migration (Week 5-6)
1. **Update Tests**
   - Move tests to appropriate packages
   - Update test imports
   - Ensure all tests pass

2. **Update Build Scripts**
   - Update Wails configuration
   - Update deployment scripts
   - Update CI/CD pipelines

3. **Documentation**
   - Update README with new structure
   - Create package documentation
   - Update development guides

## Migration Strategy

### Incremental Approach
1. **Parallel Structure**: Build new structure alongside old
2. **Gradual Migration**: Move one package at a time
3. **Maintain Compatibility**: Keep old structure working during migration
4. **Feature Flags**: Use flags to switch between old/new implementations

### Risk Mitigation
1. **Comprehensive Testing**: Full test coverage before migration
2. **Rollback Plan**: Git branches for easy rollback
3. **Staged Deployment**: Test in development before production
4. **Team Communication**: Clear communication about changes

## Benefits

### Immediate Benefits
1. **Better Code Organization**: Clear package structure
2. **Improved Navigation**: Easy to find related code
3. **Reduced Coupling**: Clear boundaries between components
4. **Enhanced Testability**: Isolated components easier to test

### Long-term Benefits
1. **Scalability**: Easy to add new features in appropriate packages
2. **Maintainability**: Clear ownership and responsibilities
3. **Onboarding**: New developers understand structure quickly
4. **Performance**: Potential for package-level optimizations

## Success Metrics

### Code Quality Metrics
- **Package Cohesion**: Average 85%+ cohesion score
- **Coupling**: No circular dependencies
- **Test Coverage**: Increase from current to 80%+
- **Code Duplication**: Reduce by 50%

### Developer Experience Metrics
- **Navigation Time**: 70% reduction in time to find code
- **Build Time**: No significant increase
- **Test Execution**: 30% faster due to better isolation
- **New Feature Time**: 40% reduction in implementation time

## Conclusion

This refactoring represents a critical investment in DealDone's technical foundation. By addressing the current architectural debt, we enable faster development, better testing, and improved maintainability. The modular structure will support the application's growth and make it easier to add new features while maintaining code quality.

The phased approach ensures minimal disruption to ongoing development while systematically improving the codebase. Success will be measured not just by the technical improvements but by the enhanced developer experience and accelerated feature delivery. 