# DealDone - Desktop File Manager

A modern, cross-platform desktop file manager built with Wails, React, and TypeScript. DealDone provides an intuitive interface for managing files and folders with native performance and a beautiful, responsive UI.

![Platform Support](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Wails](https://img.shields.io/badge/Wails-v2-red)
![React](https://img.shields.io/badge/React-18-blue)
![TypeScript](https://img.shields.io/badge/TypeScript-5-blue)

## ✨ Features

### Core File Operations
- **Tree View Navigation** - Hierarchical file system browsing with expand/collapse
- **File Management** - Create, copy, move, delete, and rename files and folders
- **System Integration** - Open files with default system applications
- **Real-time Search** - Fast file and folder search with debounced input
- **Context Menus** - Right-click operations with comprehensive file actions

### User Experience
- **Theme Support** - Light, dark, and system theme detection with seamless switching
- **Responsive Design** - Clean, modern interface built with shadcn/ui components
- **Keyboard Shortcuts** - Efficient navigation with keyboard support
- **Loading States** - Visual feedback for all operations with spinners and skeletons
- **Error Handling** - Comprehensive error boundaries with user-friendly messages

### Technical Features
- **Native Performance** - Wails framework for native desktop app experience
- **Type Safety** - Full TypeScript implementation for robust development
- **Error Recovery** - Graceful error handling with retry mechanisms
- **Accessibility** - WCAG-compliant design with proper focus management
- **Toast Notifications** - Real-time feedback for operations and status updates

## 🚀 Tech Stack

### Backend
- **Go** - High-performance backend with native file system operations
- **Wails v2** - Go + Web frontend framework for desktop applications

### Frontend
- **React 18** - Modern React with hooks and functional components
- **TypeScript 5** - Full type safety and enhanced developer experience
- **Tailwind CSS** - Utility-first styling with theme support
- **shadcn/ui** - Beautiful, accessible component library
- **Lucide React** - Consistent icon system
- **Radix UI** - Unstyled, accessible UI primitives

### Development Tools
- **Vite** - Fast build tool with hot module replacement
- **ESLint** - Code quality and consistency
- **PostCSS** - CSS processing and optimization

## 📋 Prerequisites

Before running DealDone, ensure you have the following installed:

- **Go 1.18+** - [Download Go](https://golang.org/dl/)
- **Node.js 16+** - [Download Node.js](https://nodejs.org/)
- **Wails CLI** - Install with: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Platform-Specific Requirements

**macOS:**
- Xcode Command Line Tools: `xcode-select --install`

**Windows:**
- WebView2 runtime (usually pre-installed on Windows 11)

**Linux:**
- `webkit2gtk-4.0-dev` package

## 🛠️ Installation & Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd DealDone
   ```

2. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

3. **Install frontend dependencies:**
   ```bash
   cd frontend
   npm install
   cd ..
   ```

## 🎯 Running the Application

### Development Mode

For development with hot reload:

```bash
wails dev
```

This will:
- Start the Go backend server
- Launch the React development server
- Open the desktop application window
- Enable live reloading for both frontend and backend changes
- Provide access to browser dev tools at `http://localhost:34115`

### Production Build

To create an optimized production build:

```bash
wails build
```

The executable will be created in `build/bin/` directory.

### Platform-Specific Builds

```bash
# macOS (Intel)
wails build -platform darwin/amd64

# macOS (Apple Silicon)
wails build -platform darwin/arm64

# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64
```

## 📖 Usage Guide

### Basic Navigation
- **Tree Navigation** - Click folders to expand/collapse directory trees
- **File Selection** - Click files to select them
- **Context Menu** - Right-click any file or folder for available actions

### File Operations
- **Create** - Right-click in a folder → "New File" or "New Folder"
- **Copy/Cut** - Right-click item → "Copy" or "Cut", then paste in destination
- **Rename** - Right-click item → "Rename" and enter new name
- **Delete** - Right-click item → "Delete" (confirmation dialog will appear)
- **Open** - Double-click files or right-click → "Open"

### Keyboard Shortcuts
- `Cmd/Ctrl + /` - Show keyboard shortcuts help
- `F5` - Refresh current directory
- `Escape` - Cancel current operation or close dialogs
- `Enter` - Confirm dialog actions
- `Delete` - Delete selected items

### Theme Switching
Click the theme toggle in the top-right corner to switch between:
- **Light** - Light theme
- **Dark** - Dark theme  
- **System** - Follow system preference

## 📁 Project Structure

```
DealDone/
├── frontend/                 # React frontend application
│   ├── src/
│   │   ├── components/      # React components
│   │   │   ├── ui/         # shadcn/ui components
│   │   │   ├── FileTree.tsx
│   │   │   ├── ContextMenu.tsx
│   │   │   └── ...
│   │   ├── contexts/       # React contexts
│   │   ├── hooks/          # Custom React hooks
│   │   ├── services/       # API services
│   │   ├── types/          # TypeScript type definitions
│   │   └── utils/          # Utility functions
│   ├── wailsjs/           # Auto-generated Wails bindings
│   └── ...
├── tasks/                  # Project documentation
├── *.go                   # Go backend files
├── wails.json            # Wails configuration
└── README.md
```

## 🔧 Development

### Adding New Features

1. **Backend (Go)** - Add methods to `app.go` and implement in appropriate files
2. **Frontend (React)** - Create components in `src/components/`
3. **Types** - Update TypeScript types in `src/types/`
4. **API Integration** - Update `src/services/fileManagerApi.ts`

### Code Style

- **Go** - Follow standard Go conventions
- **TypeScript/React** - Use functional components with hooks
- **Styling** - Use Tailwind CSS classes and shadcn/ui components
- **Error Handling** - Use the centralized `ErrorService`

### Testing

Currently focused on manual testing. Automated testing setup can be added with:
- **Go** - Standard `go test`
- **Frontend** - Jest + React Testing Library

## 🐛 Troubleshooting

### Common Issues

**Port Conflicts:**
- Wails dev server uses port 34115 by default
- Change in `wails.json` if needed

**Build Errors:**
- Ensure all dependencies are installed: `go mod tidy` and `npm install`
- Check Go and Node.js versions match requirements

**Permission Issues (macOS):**
- Allow app in System Preferences → Security & Privacy
- For development builds: `sudo spctl --master-disable`

**Performance:**
- Large directories may load slowly - this is expected behavior
- Consider pagination for directories with 1000+ items

## 📝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Wails](https://wails.io/) - Go + Web frontend framework
- [shadcn/ui](https://ui.shadcn.com/) - Beautiful component library
- [Lucide](https://lucide.dev/) - Icon system
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework

---

Built with ❤️ using Wails, React, and Go
