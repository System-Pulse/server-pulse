# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Server Pulse is a terminal-based system monitoring tool built with Go and the Bubbletea TUI framework. It provides real-time monitoring of system resources, Docker containers, network diagnostics, security checks, and performance analysis.

## Build and Development Commands

### Building
```bash
# Build for current platform
make build

# Build for all platforms (creates _build directory)
make build-all

# Clean build artifacts
make clean
```

### Running
```bash
# Run in development mode with debugging
make run-dev

# Run directly
./server-pulse
```

### Testing
```bash
# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./system/app/test/
go test -v ./system/process/test/
go test -v ./system/resource/test/
go test -v ./system/security/test/
```

## Architecture Overview

### High-Level Structure

Server Pulse follows a clean separation between **system monitoring** and **UI presentation**:

1. **main.go**: Entry point that initializes Docker manager and runs the TUI loop with support for dropping into interactive shells
2. **system/**: Backend packages for system monitoring and data collection
3. **widgets/**: Frontend TUI components using Bubbletea/Lipgloss
4. **utils/**: Shared utilities (Docker permissions, etc.)

### The Bubbletea Pattern

This project uses the Elm Architecture via Bubbletea:
- **Model** ([widgets/model.go](widgets/model.go)): Central state container with nested models for different views
- **Init** ([widgets/init.go](widgets/init.go)): Initialization that creates tables, inputs, and managers
- **Update** ([widgets/handles.go](widgets/handles.go), [widgets/handles-common.go](widgets/handles-common.go)): Message handling for state transitions
- **View** ([widgets/view.go](widgets/view.go), [widgets/render-*.go](widgets/render-body.go)): Rendering logic organized by component

### Key System Packages

**system/app/**: Docker container management
- Uses Docker client API to list, inspect, control containers
- [operations.go](system/app/operations.go) defines container operations (start/stop/pause/restart/delete)
- [app.go](system/app/app.go) provides Bubbletea commands that return messages
- Supports interactive shell execution via `ExecInteractiveShellAlternative`

**system/resource/**: System resource monitoring
- CPU, memory, disk, and network statistics
- Uses gopsutil/v4 for cross-platform system info

**system/process/**: Process listing and management
- Retrieves running processes with CPU/memory usage

**system/security/**: Security diagnostics
- SSL certificate checks
- SSH configuration auditing (root login, password auth)
- Firewall status (ufw, iptables)
- Password policy checks
- Auto-ban systems (fail2ban)
- System update checks
- Requires root or sudo for most checks

**system/performance/**: Performance analysis
- System health scoring based on metrics
- I/O wait, context switches, interrupts monitoring
- CPU and memory benchmarks
- [systemHealth.go](system/performance/systemHealth.go) calculates health scores

**system/network/**: Network diagnostics
- Interface statistics
- Active connections
- Routing tables
- DNS configuration
- Ping and traceroute functionality

**system/logs/**: System log analysis
- Journalctl integration
- Log filtering by time range, service, level
- Structured log parsing

### State Management

The application uses a state machine defined in [widgets/model/ui.go](widgets/model/ui.go):
- States: `StateHome`, `StateMonitor`, `StateSystem`, `StateProcess`, `StateContainers`, `StateContainer`, `StateContainerLogs`, `StateDiagnostics`, `StateNetwork`, `StateReporting`, etc.
- Navigation uses `setState()` and `goBack()` methods in [widgets/model.go](widgets/model.go)
- Tab-based navigation with states mapped to specific tabs

### Authentication Flow

Many operations require elevated privileges:
- Checks if running as root via `utils.IsRoot()`
- Falls back to sudo if available (`canRunSudo()`)
- Password input widget for sudo operations
- Authentication state tracked in models (e.g., `Network.AuthState`, `Diagnostic.AuthState`)
- Uses `AuthNotRequired`, `AuthRequired`, `AuthSuccess`, `AuthFailed` states

### Container Interactive Shell Flow

Container shell execution has special handling:
1. User requests shell in container
2. Model stores `PendingShellExec` with container ID
3. TUI exits cleanly
4. main.go detects pending shell and calls `ExecInteractiveShellAlternative`
5. User interacts with shell
6. Shell exits, TUI restarts with preserved state

### Message-Based Updates

All async operations return Bubbletea commands that produce messages:
- `ContainerMsg`: Updated container list
- `ContainerOperationMsg`: Result of container operations
- `SecurityMsg`: Security check results
- `TickMsg`: Timer for periodic updates
- Messages defined across system packages (e.g., [system/app/model.go](system/app/model.go))

### Rendering Architecture

Rendering is split across multiple files:
- [widgets/render-header.go](widgets/render-header.go): Top bar with title and system info
- [widgets/render-nav.go](widgets/render-nav.go): Tab navigation
- [widgets/render-body.go](widgets/render-body.go): Main content area dispatcher
- [widgets/render-monitor.go](widgets/render-monitor.go): Resource monitoring views
- [widgets/render-network.go](widgets/render-network.go): Network diagnostics views
- [widgets/render-diagnostic.go](widgets/render-diagnostic.go): Security and performance views
- [widgets/render-table.go](widgets/render-table.go): Table rendering helpers
- [widgets/render-charts.go](widgets/render-charts.go): ASCII graph rendering using asciigraph

### Testing Strategy

Tests use table-driven patterns and mocks:
- [system/app/test/mocks.go](system/app/test/mocks.go) provides Docker client mocks
- Tests verify message handling and data transformations
- Some tests use `t.Parallel()` for concurrent execution

## Development Guidelines

### Adding New System Monitoring Features

1. Create functionality in appropriate `system/` package
2. Define message types for Bubbletea in package (e.g., `type MyDataMsg struct{}`)
3. Create tea.Cmd function that returns the message
4. Add message handler in `widgets/handles.go` Update() method
5. Add rendering logic in appropriate `widgets/render-*.go` file
6. Update model state in `widgets/model/` if needed

### Adding New UI Views

1. Define new state constant in [widgets/model/ui.go](widgets/model/ui.go)
2. Add state transition logic in `setState()` and `goBack()` in [widgets/model.go](widgets/model.go)
3. Create rendering function in new or existing `widgets/render-*.go` file
4. Add keyboard handling in [widgets/handles.go](widgets/handles.go) or [widgets/handles-common.go](widgets/handles-common.go)
5. Update navigation tabs if needed in [widgets/vars/vars.go](widgets/vars/vars.go)

### Working with Tables

Tables use charmbracelet/bubbles table component:
- Define columns in init.go with Title and Width
- Set custom styles using `table.DefaultStyles()`
- Update rows dynamically via `table.SetRows()`
- Handle selection in Update() method
- Render with `table.View()`

### Performance Considerations

- Chart updates throttled via `LastChartUpdate` to avoid excessive rendering
- Data histories limited via `MaxPoints` (typically 60 points)
- Container stats updates run in goroutines with message passing
- Animations can be disabled via `EnableAnimations` flag

### Security Feature Requirements

Most security checks require elevated permissions:
- Always check `IsRoot` or `CanRunSudo` before attempting operations
- Use `SecurityManager` for coordinating security checks
- Handle authentication failures gracefully with user-friendly messages
- Never store passwords permanently (only in memory during session)
