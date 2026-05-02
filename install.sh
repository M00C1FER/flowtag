#!/usr/bin/env bash
# flowtag — interactive install wizard.
set -euo pipefail

if [ -t 1 ]; then C_BOLD="$(tput bold)"; C_RESET="$(tput sgr0)"; C_GREEN="$(tput setaf 2)"; C_YELLOW="$(tput setaf 3)"; C_RED="$(tput setaf 1)"; else C_BOLD=""; C_RESET=""; C_GREEN=""; C_YELLOW=""; C_RED=""; fi
say()  { printf "%s%s%s\n" "$C_BOLD" "$1" "$C_RESET"; }
info() { printf "  %s\n" "$1"; }
ok()   { printf "  %s✓%s %s\n" "$C_GREEN" "$C_RESET" "$1"; }
warn() { printf "  %s!%s %s\n" "$C_YELLOW" "$C_RESET" "$1"; }
fail() { printf "  %s✗%s %s\n" "$C_RED" "$C_RESET" "$1" >&2; exit 1; }
prompt_yn() { local q="$1" def="${2:-y}" ans; if [ "$def" = "y" ]; then read -r -p "  $q [Y/n]: " ans; ans="${ans:-y}"; else read -r -p "  $q [y/N]: " ans; ans="${ans:-n}"; fi; [[ "$ans" =~ ^[Yy] ]]; }
prompt_default() { read -r -p "  $1 [$2]: " ans; echo "${ans:-$2}"; }

detect_os() { OS_ID=unknown; OS_LIKE=""; OS_VERSION=""; OS_WSL=0; OS_TERMUX=0; [ -f /etc/os-release ] && { . /etc/os-release; OS_ID="${ID:-}"; OS_LIKE="${ID_LIKE:-}"; OS_VERSION="${VERSION_ID:-}"; }; [ "$(uname)" = "Darwin" ] && OS_ID=macos; grep -qi microsoft /proc/sys/kernel/osrelease 2>/dev/null && OS_WSL=1 || true; [ -n "${TERMUX_VERSION:-}" ] && OS_TERMUX=1 && OS_ID=termux; }
pkg_install() {
    case "$OS_ID" in
        debian|ubuntu) sudo apt-get update -qq && sudo apt-get install -y "$@";;
        fedora|rhel|centos) sudo dnf install -y "$@";;
        arch|manjaro) sudo pacman -S --noconfirm "$@";;
        alpine) sudo apk add --no-cache "$@";;
        opensuse*|sles) sudo zypper install -y "$@";;
        macos) brew install "$@";;
        termux) pkg install -y "$@";;
        *) warn "unknown OS — install manually: $*"; return 1;;
    esac
}
ensure_go() {
    command -v go >/dev/null && { ok "Go: $(go version | awk '{print $3}')"; return 0; }
    if prompt_yn "Install Go via system package manager?"; then
        pkg_install go || pkg_install golang || pkg_install golang-go || fail "Go install failed"
    else fail "Go 1.22+ required"; fi
}

main() {
    say "flowtag — install wizard (single Go binary)"
    detect_os
    info "OS: ${OS_ID}${OS_VERSION:+ $OS_VERSION}$([ "$OS_WSL" = 1 ] && echo ' (WSL2)')$([ "$OS_TERMUX" = 1 ] && echo ' (Termux)')"

    say ""; say "Step 1/3: Go toolchain"; ensure_go
    say ""; say "Step 2/3: Install"
    local BIN_DIR
    BIN_DIR="$(prompt_default "Binary directory (must be in \$PATH)" "$HOME/.local/bin")"
    mkdir -p "$BIN_DIR"
    if prompt_yn "Install via 'go install' (recommended)?" y; then
        GOBIN="$BIN_DIR" go install github.com/M00C1FER/flowtag/cmd/flowtag@latest
    else
        local INSTALL_HOME; INSTALL_HOME="$(prompt_default "Source checkout root" "$HOME/.local/share/flowtag")"
        mkdir -p "$INSTALL_HOME"
        if [ -d "$INSTALL_HOME/.git" ]; then ( cd "$INSTALL_HOME" && git pull -q ); else git clone -q https://github.com/M00C1FER/flowtag.git "$INSTALL_HOME"; fi
        ( cd "$INSTALL_HOME" && go build -o "$BIN_DIR/flowtag" ./cmd/flowtag )
    fi
    ok "binary at $BIN_DIR/flowtag"

    say ""; say "Step 3/3: Verify"
    if "$BIN_DIR/flowtag" --help 2>&1 | head -1 | grep -qi "flowtag"; then ok "flowtag --help works"; else warn "verification failed"; fi
    say ""
    ok "Done. In a repo with conventional commits, try: flowtag --next-version"
}
main "$@"
