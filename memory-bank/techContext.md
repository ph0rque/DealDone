# Technical Context: DealDone

## Technology Stack

### Desktop Framework
**Wails v2**
- Cross-platform desktop apps with Go backend and web frontend
- Native OS integration for file system operations
- Built-in IPC for frontend-backend communication
- Supports Windows, macOS, and Linux

### Backend Technologies
**Go 1.21+**
- Primary backend language
- Libraries:
  - `fsnotify`: File system event monitoring
  - `go-chi/chi`: HTTP router for API endpoints
  - `stretchr/testify`: Testing framework
  - `golang/mock`: Mocking for unit tests

### Frontend Technologies
**React 18 + TypeScript**
- Component-based UI architecture
- Type safety and better developer experience
- Libraries:
  - `react-router-dom`: Client-side routing
  - `axios`: HTTP client for API calls
  - `react-query`: Server state management
  - `react-hook-form`: Form handling

**Tailwind CSS**
- Utility-first CSS framework
- Responsive design out of the box
- Custom component styling

**Vite**
- Fast build tool and dev server
- HMR (Hot Module Replacement)
- Optimized production builds

### AI Integration
**n8n (Self-hosted or Cloud)**
- Workflow automation platform
- Connects desktop app to AI services
- Handles async processing pipelines

**Anthropic Claude API**
- Document classification
- Data extraction from documents
- Conversational AI for chat interface
- Model: Claude 3 Opus/Sonnet

### Development Setup

#### Prerequisites
```bash
# Go 1.21+
go version

# Node.js 18+
node --version

# Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### Project Structure
```
DealDone/
├── app.go              # Wails app configuration
├── main.go             # Entry point
├── frontend/           # React application
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── internal/           # Go internal packages
├── build/              # Build configuration
└── wails.json          # Wails configuration
```

#### Build Commands
```bash
# Development
wails dev

# Production build
wails build -platform darwin/universal
wails build -platform windows/amd64
wails build -platform linux/amd64
```

## External Dependencies

### Runtime Dependencies
- **File System Access**: Full read/write permissions in user directory
- **Network Access**: For n8n API communication
- **OS Integration**: Native file dialogs and system notifications

### API Dependencies
- **n8n Instance**: Self-hosted or cloud
- **Anthropic API Key**: For Claude access
- **Optional**: OCR service for scanned documents

## Configuration

### Environment Variables
```env
# .env.local (frontend)
VITE_API_BASE_URL=http://localhost:3000
VITE_N8N_WEBHOOK_URL=https://your-n8n.com/webhook

# Go environment
DEALDONE_DATA_DIR=$HOME/DealDone
N8N_API_KEY=your-api-key
ANTHROPIC_API_KEY=your-claude-key
```

### Wails Configuration
```json
{
  "name": "DealDone",
  "outputfilename": "DealDone",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "author": {
    "name": "DealDone Team",
    "email": "team@dealdone.com"
  }
}
```

## Development Workflow

### Local Development
1. Run `wails dev` for live reload
2. Frontend on `http://localhost:5173`
3. Go backend auto-recompiles on changes
4. Wails provides dev tools in app window

### Testing Strategy
- **Frontend**: Jest + React Testing Library
- **Backend**: Go standard testing + testify
- **Integration**: Wails provides testing utilities
- **E2E**: Playwright for desktop app testing

### CI/CD Pipeline
```yaml
# GitHub Actions example
- Build for all platforms
- Run unit tests
- Run integration tests
- Code signing for distribution
- Auto-update system integration
```

## Performance Considerations

### Frontend Optimization
- Code splitting for lazy loading
- Memoization for expensive computations
- Virtual scrolling for large document lists
- Web Workers for heavy processing

### Backend Optimization
- Goroutines for concurrent file processing
- Channel-based job queuing
- Connection pooling for API calls
- Efficient file streaming

## Security Considerations

### Application Security
- Code signing certificates for distribution
- Sandboxed file operations
- Input sanitization
- XSS prevention in React

### Data Security
- Local encryption for sensitive data
- Secure API key storage (OS keychain)
- HTTPS for all external communication
- No telemetry without consent 