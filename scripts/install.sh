#!/bin/bash
# scripts/install.sh - Install ai-dev-tools
# Usage: curl -sfL https://get.ai-dev.tools | sh
#        or: curl -sfL https://raw.githubusercontent.com/ernedo/ai-dev-tools/main/scripts/install.sh | sh

set -euo pipefail

REPO="ernedo/ai-dev-tools"
BINARY="ai-dev"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "$OS" in
        darwin) OS="Darwin" ;;
        linux) OS="Linux" ;;
        *) log_error "Unsupported OS: $OS"; exit 1 ;;
    esac

    case "$ARCH" in
        x86_64) ARCH="x86_64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    log_info "Detected platform: $OS/$ARCH"
}

# Get latest release version
get_latest_version() {
    log_info "Fetching latest release..."
    VERSION=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    if [ -z "$VERSION" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
    log_info "Latest version: $VERSION"
}

# Download and install binary
install_binary() {
    local asset_name="ai-dev-tools_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    local url="https://github.com/$REPO/releases/download/$VERSION/$asset_name"
    local tmp_dir=$(mktemp -d)
    local tmp_file="$tmp_dir/$asset_name"

    log_info "Downloading $asset_name..."
    if ! curl -sfL "$url" -o "$tmp_file"; then
        log_error "Failed to download $url"
        log_warn "Try installing via: brew tap ernedo/tap && brew install ai-dev-tools"
        log_warn "Or: npm i -g @ernedo/ai-dev-tools"
        exit 1
    fi

    log_info "Extracting..."
    tar -xzf "$tmp_file" -C "$tmp_dir"

    log_info "Installing to $INSTALL_DIR..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp_dir/$BINARY" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY"
    else
        log_warn "Need sudo to install to $INSTALL_DIR"
        sudo mv "$tmp_dir/$BINARY" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY"
    fi

    rm -rf "$tmp_dir"
}

# Verify installation
verify_install() {
    if command -v "$BINARY" >/dev/null 2>&1; then
        local version=$("$BINARY" version 2>/dev/null | head -1 || echo "unknown")
        log_info "Successfully installed: $version"
    else
        log_error "Installation failed: $BINARY not found in PATH"
        exit 1
    fi
}

main() {
    log_info "Installing ai-dev-tools..."
    detect_platform
    get_latest_version
    install_binary
    verify_install

    echo
    log_info "Next steps:"
    echo "  1. Configure your LLM provider: $BINARY config init"
    echo "  2. Generate a commit: $BINARY commit --dry-run"
    echo "  3. Review code: $BINARY review --help"
    echo "  4. Generate tests: $BINARY testgen --help"
    echo "  5. Create changelog: $BINARY changelog --help"
}

main "$@"