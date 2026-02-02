<p align="center"><img width="200px" src="docs/images/logo.png" alt="server-pulse"/></p>

# Server Pulse 🚀

![License: GPL-3.0](https://img.shields.io/badge/license-GPL3.0-blue)
![release](https://img.shields.io/github/v/release/System-Pulse/server-pulse)

A real-time Terminal User Interface (TUI) for system monitoring, Docker container management, network diagnostics, and security auditing on Linux servers.

<p align="center"><img src="docs/images/sample.gif" alt="server-pulse"/></p>

## Features

- **System Monitoring** — Real-time CPU, memory, disk, and network usage with visual graphs
- **Process Management** — List, search, sort, and kill processes
- **Docker Management** — Start, stop, restart, pause, remove containers; view logs (with live streaming); exec into shells; monitor per-container CPU/memory/network stats
- **Security Diagnostics** — SSL certificate checks, SSH root access audit, open port scanning, firewall rules, Fail2Ban status
- **Performance Analysis** — System health scoring, I/O metrics, per-core CPU breakdown, memory analysis
- **Log Viewer** — Filter system logs by time range, service, and log level
- **Network Diagnostics** — Interface stats, ping, traceroute, speed tests, routing table, DNS servers, connection analysis
- **Report Generation** — Generate, save, and load comprehensive system health reports (`~/.server-pulse/reports/`)

## Requirements

- **Linux** (uses `/proc` filesystem — no macOS or Windows support)
- **Docker** installed and running
- Current user must be in the `docker` group or run as root
- Terminal size of at least 80x20

## Installation

### Pre-built Binaries (Recommended)

**curl:**

```bash
curl -fsSL https://raw.githubusercontent.com/System-Pulse/server-pulse/main/scripts/install.sh | bash
```

**wget:**

```bash
wget -qO- https://raw.githubusercontent.com/System-Pulse/server-pulse/main/scripts/install.sh | bash
```

The install script auto-detects your platform, downloads the latest release, verifies the SHA256 checksum, and installs the binary to `/usr/local/bin/`.

### Manual Download

Download a specific release for Linux amd64:

```bash
sudo wget https://github.com/System-Pulse/server-pulse/releases/download/vX.X.X/server-pulse-X.X.X-linux-amd64 -O /usr/local/bin/server-pulse
sudo chmod +x /usr/local/bin/server-pulse
```

### From Source

#### Prerequisites
- Go 1.24+
- Git

```bash
git clone https://github.com/System-Pulse/server-pulse.git
cd server-pulse
make build
sudo mv server-pulse /usr/local/bin/
```

## Usage

After installation, launch the TUI:

```bash
server-pulse
```

The application opens an interactive dashboard. Use the number keys to jump between the four main sections:

| Key | Section | Description |
|-----|---------|-------------|
| `1` | Monitor | System resources, processes, and containers |
| `2` | Diagnostics | Security checks, performance analysis, and logs |
| `3` | Network | Interfaces, connectivity tests, routing, and protocol analysis |
| `4` | Reporting | Generate and manage system health reports |

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `?` | Toggle help |
| `q` / `Esc` / `Ctrl+C` | Quit |
| `b` | Go back |
| `Tab` / `←` `→` | Navigate tabs |
| `Enter` | Select |

### Process View

| Key | Action |
|-----|--------|
| `/` | Search |
| `s` | Sort by CPU |
| `m` | Sort by memory |
| `k` | Kill process |

### Containers View

| Key | Action |
|-----|--------|
| `/` | Search |
| `Enter` | Open container actions (details, logs, restart, delete, exec shell, etc.) |

### Container Logs

| Key | Action |
|-----|--------|
| `s` | Toggle live streaming |
| `r` | Refresh |
| `Home` / `End` | Jump to start/end |

### Network View

| Key | Action |
|-----|--------|
| `p` | Start ping |
| `t` | Start traceroute |
| `c` | Clear results |

### Diagnostics View

| Key | Action |
|-----|--------|
| `1` `2` `3` | Switch sub-tabs (Security / Performance / Logs) |
| `/` | Search |
| `r` | Reload |
| `Shift+←` `Shift+→` | Switch filters |

### Reporting View

| Key | Action |
|-----|--------|
| `g` | Generate report |
| `s` | Save report |
| `l` | Load saved reports |
| `d` | Delete report |

## Uninstallation

**curl:**

```bash
curl -fsSL https://raw.githubusercontent.com/System-Pulse/server-pulse/main/scripts/uninstall.sh | bash
```

**wget:**

```bash
wget -qO- https://raw.githubusercontent.com/System-Pulse/server-pulse/main/scripts/uninstall.sh | bash
```

## Building

1. Clone the repository:
```bash
git clone https://github.com/System-Pulse/server-pulse.git
cd server-pulse
```

2. Build for your current platform:
```bash
make build
```

3. For cross-platform builds:
```bash
make build-all
```

Build artifacts will be placed in the `_build` directory.

### Supported Build Targets

| Architecture | Platform |
|-------------|----------|
| amd64 | Linux |
| arm | Linux |
| arm64 | Linux |
| ppc64le | Linux |

## License

This project is licensed under the [GPL-3.0 License](LICENSE).
