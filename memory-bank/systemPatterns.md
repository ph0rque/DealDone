# System Patterns: DealDone

## Architecture Overview

### Desktop Application (Wails)
```
┌─────────────────────────────────────────┐
│           Frontend (React/TS)           │
│  ┌─────────────┐  ┌─────────────────┐  │
│  │   UI Layer  │  │  State Manager  │  │
│  └─────────────┘  └─────────────────┘  │
└────────────────────┬────────────────────┘
                     │ Wails Bridge
┌────────────────────┴────────────────────┐
│           Backend (Go)                  │
│  ┌─────────────┐  ┌─────────────────┐  │
│  │File Manager │  │  API Gateway    │  │
│  └─────────────┘  └─────────────────┘  │
└─────────────────────────────────────────┘
                     │
                     ├── Local File System
                     └── n8n API
```

## Core Patterns

### 1. Event-Driven Architecture
- File system events trigger processing pipelines
- User actions emit events for state updates
- Background workers handle async operations

### 2. Repository Pattern
```typescript
interface DocumentRepository {
  categorize(document: Document): Promise<Category>
  extract(document: Document): Promise<ExtractedData>
  save(document: Document, category: Category): Promise<void>
}
```

### 3. Strategy Pattern for Document Processing
```go
type DocumentProcessor interface {
    Process(doc Document) (ProcessedDocument, error)
}

type LegalProcessor struct{}
type FinancialProcessor struct{}
type GeneralProcessor struct{}
```

### 4. Observer Pattern for Learning System
- Monitors user corrections in analysis files
- Notifies AI training pipeline of changes
- Updates confidence models based on feedback

## Component Architecture

### Frontend Components
```
src/
├── components/
│   ├── DocumentDropZone/     # Drag-and-drop interface
│   ├── CategoryView/         # Document category display
│   ├── AnalysisPanel/        # Template data view
│   ├── ConfidenceIndicator/  # Data confidence display
│   └── AIChat/              # Conversational interface
├── contexts/
│   ├── DocumentContext      # Document state management
│   └── AnalysisContext      # Analysis data state
└── services/
    ├── documentService      # Document operations
    └── aiService           # AI integration
```

### Backend Services
```
internal/
├── filemanager/     # File system operations
├── categorizer/     # Document classification
├── extractor/       # Data extraction logic
├── analyzer/        # Template population
└── api/            # n8n communication
```

## Data Flow Patterns

### Document Processing Pipeline
1. **Ingestion**: File dropped → Event emitted
2. **Classification**: Content analyzed → Category determined
3. **Storage**: File moved → Database updated
4. **Extraction**: AI processes → Data extracted
5. **Population**: Template filled → Confidence scored
6. **Notification**: UI updated → User informed

### State Management Flow
```
User Action → Frontend Event → Context Update → UI Re-render
     ↓
Backend API ← Wails Bridge ← Service Call
     ↓
File System / n8n API
```

## Error Handling Patterns

### Graceful Degradation
- If AI unavailable: Queue for later processing
- If categorization fails: Place in "uncategorized"
- If extraction fails: Flag for manual review

### Retry Strategy
```go
type RetryConfig struct {
    MaxAttempts int
    BackoffMs   int
    Multiplier  float64
}
```

## Security Patterns

### Sandboxed File Operations
- All file operations within DealDone directory
- No execution of uploaded files
- Sanitized file paths

### API Security
- Encrypted communication with n8n
- API key rotation support
- Request rate limiting

## Performance Patterns

### Lazy Loading
- Documents loaded on demand
- Pagination for large document sets
- Progressive rendering of analysis results

### Caching Strategy
- LRU cache for recent documents
- Memoized AI responses
- Persistent cache for offline mode 