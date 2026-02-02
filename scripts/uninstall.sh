#!/usr/bin/env bash
# Simple uninstallation script for server-pulse

# Output functions
function output() { echo -e "[server-pulse-uninstall] $*"; }
function log_warning() { echo -e "[WARNING] $1"; }
function log_error() { echo -e "[ERROR] $1"; }

# Check if command exists
function command_exists() {
    command -v "$@" > /dev/null 2>&1
}

# Determine if we need sudo/su
sh_c='sh -c'
if [[ $EUID -ne 0 ]]; then
    if command_exists sudo; then
        log_warning "sudo is required to uninstall server-pulse"
        sh_c='sudo -E sh -c'
    elif command_exists su; then
        log_warning "su is required to uninstall server-pulse"
        sh_c='su -c'
    else
        log_error "This uninstaller needs root privileges. Neither 'sudo' nor 'su' were found."
        exit 1
    fi
fi

# Default installation path
BIN_PATH="/usr/local/bin/server-pulse"

output "Checking for server-pulse installation..."
if [ -f "$BIN_PATH" ]; then
    output "Found server-pulse at $BIN_PATH"

    output "Removing server-pulse..."
    if $sh_c "rm -f $BIN_PATH"; then
        output "server-pulse was successfully uninstalled"
    else
        log_error "Failed to remove server-pulse"
        exit 1
    fi
else
    output "server-pulse is not installed at $BIN_PATH"
    
    # Check if it's installed elsewhere
    alternative_path=$(command -v server-pulse 2>/dev/null)
    if [ -n "$alternative_path" ]; then
        output "Found server-pulse at $alternative_path"
        read -p "Do you want to remove it? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if $sh_c "rm -f $alternative_path"; then
                output "server-pulse was successfully uninstalled"
            else
                log_error "Failed to remove server-pulse"
                exit 1
            fi
        else
            output "Uninstallation cancelled"
        fi
    else
        output "No server-pulse installation found"
    fi
fi

output "Uninstallation complete"
