#!/usr/bin/env bash
# a simple install script for server-pulse

KERNEL=$(uname -s)


function log_warning() { 
    echo -e "[WARNING] $1" 
}
function log_error() { 
    echo -e "[ERROR] $1"
}
function output() { 
    echo -e "[server-pulse-install] $*"
}

function command_exists() {
  command -v "$@" > /dev/null 2>&1
}

# extract github download url matching pattern
function extract_url() {
  local match=$1
  shift
  while read -r line; do
    case $line in
      *browser_download_url*"${match}"*)
        url=$(echo "$line" | sed -e 's/^.*"browser_download_url":[ ]*"//' -e 's/".*//;s/\ //g')
        echo "$url"
        break
      ;;
    esac
  done <<< "$*"
}

case $KERNEL in
  Linux) MATCH_BUILD="linux-amd64" ;;
  Darwin) MATCH_BUILD="darwin-amd64" ;;
  *)
    log_error "Platform not supported by this install script"
    exit 1
    ;;
esac

for req in curl wget; do
  command_exists "$req" || {
    output "Missing required $req binary"
    req_failed=1
  }
done
[ "$req_failed" = 1 ] && exit 1

sh_c='sh -c'
if [[ $EUID -ne 0 ]]; then
  if command_exists sudo; then
    log_warning "sudo is required to install server-pulse"
    sh_c='sudo -E sh -c'
  elif command_exists su; then
    log_warning "su is required to install server-pulse"
    sh_c='su -c'
  else
    log_error "This installer needs the ability to run commands as root. We are unable to find either sudo or su available to make this happen."
    exit 1
  fi
fi

TMP=$(mktemp -d "${TMPDIR:-/tmp}/server-pulse.XXXXX")
cd "${TMP}" || exit

output "Fetching latest release info"
resp=$(curl -s https://api.github.com/repos/System-Pulse/server-pulse/releases/latest)

output "Fetching release checksums"
checksum_url=$(extract_url sha256sums.txt "$resp")
wget -q "$checksum_url" -O sha256sums.txt

# skip if latest already installed
cur_server_pulse=$(command -v server-pulse 2> /dev/null)
if [[ -n "$cur_server_pulse" ]]; then
  cur_sum=$(sha256sum "$cur_server_pulse" | sed 's/ .*//')
  (grep -q "$cur_sum" sha256sums.txt) && {
    output "server-pulse is already up-to-date"
    exit 0
  }
fi

output "Fetching latest server-pulse"
url=$(extract_url "$MATCH_BUILD" "$resp")
wget -q --show-progress "$url"
(sha256sum -c --quiet --ignore-missing sha256sums.txt) || exit 1

output "Installing to /usr/local/bin"
chmod +x server-pulse-*
$sh_c "mv server-pulse-* /usr/local/bin/server-pulse"

output "Installation complete!"
