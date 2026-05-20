#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# duck-ai — Install Script
# Personal Claude Code toolkit — skills, commands, and setup scripts.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/educlopez/duck-ai/main/install.sh | bash
#
# Pin a specific version:
#   DUCK_AI_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/educlopez/duck-ai/main/install.sh | bash
#
# Override install directory:
#   DUCK_AI_INSTALL_DIR=/usr/local/bin curl -fsSL ... | bash
# ============================================================================

GITHUB_OWNER="educlopez"
GITHUB_REPO="duck-ai"
BINARY_NAME="duck-ai"

# ============================================================================
# Color support
# ============================================================================

setup_colors() {
    if [ -t 1 ] && [ "${TERM:-}" != "dumb" ]; then
        RED='\033[0;31m'
        GREEN='\033[0;32m'
        YELLOW='\033[1;33m'
        BLUE='\033[0;34m'
        CYAN='\033[0;36m'
        BOLD='\033[1m'
        DIM='\033[2m'
        NC='\033[0m'
    else
        RED='' GREEN='' YELLOW='' BLUE='' CYAN='' BOLD='' DIM='' NC=''
    fi
}

info()    { printf '%b[info]%b    %s\n' "$BLUE" "$NC" "$*"; }
success() { printf '%b[ok]%b      %s\n' "$GREEN" "$NC" "$*"; }
warn()    { printf '%b[warn]%b    %s\n' "$YELLOW" "$NC" "$*"; }
error()   { printf '%b[error]%b   %s\n' "$RED" "$NC" "$*" >&2; }
fatal()   { error "$@"; exit 1; }
step()    { printf '\n%b%b==>%b %b%s%b\n' "$CYAN" "$BOLD" "$NC" "$BOLD" "$*" "$NC"; }

# ============================================================================
# Platform detection
# ============================================================================

detect_platform() {
    local uname_os uname_arch

    uname_os="$(uname -s)"
    uname_arch="$(uname -m)"

    case "$uname_os" in
        Darwin) OS="darwin"; OS_LABEL="macOS" ;;
        Linux)  OS="linux";  OS_LABEL="Linux" ;;
        *)      fatal "Unsupported OS: $uname_os. Only macOS and Linux are supported." ;;
    esac

    case "$uname_arch" in
        x86_64|amd64)   ARCH="amd64" ;;
        arm64|aarch64)  ARCH="arm64" ;;
        *)              fatal "Unsupported architecture: $uname_arch. Only amd64 and arm64 are supported." ;;
    esac

    success "Platform: ${OS_LABEL} (${OS}/${ARCH})"
}

# ============================================================================
# Prerequisites
# ============================================================================

check_prerequisites() {
    local missing=()

    command -v curl >/dev/null 2>&1 || missing+=("curl")
    command -v tar  >/dev/null 2>&1 || missing+=("tar")

    if [ ${#missing[@]} -gt 0 ]; then
        fatal "Missing required tools: ${missing[*]}. Please install them and try again."
    fi
}

# ============================================================================
# Version resolution
# ============================================================================

resolve_version() {
    if [ -n "${DUCK_AI_VERSION:-}" ]; then
        VERSION_TAG="$DUCK_AI_VERSION"
        # Allow user to pass either "v0.2.0" or "0.2.0"
        case "$VERSION_TAG" in
            v*) ;;
            *)  VERSION_TAG="v${VERSION_TAG}" ;;
        esac
        info "Using pinned version: ${VERSION_TAG}"
    else
        info "Fetching latest release from GitHub..."
        local url="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest"
        local response http_code body
        response="$(curl -sL -w '\n%{http_code}' "$url")" \
            || fatal "Failed to query GitHub API"
        http_code="$(printf '%s\n' "$response" | tail -n 1)"
        body="$(printf '%s\n' "$response" | sed '$d')"

        if [ "$http_code" != "200" ]; then
            fatal "GitHub API returned HTTP $http_code. Rate limited? Try again or pin DUCK_AI_VERSION."
        fi

        # Extract tag_name — works without jq
        VERSION_TAG="$(printf '%s' "$body" \
            | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
            | head -n 1)"

        if [ -z "$VERSION_TAG" ]; then
            fatal "Could not determine latest version from GitHub API response."
        fi
        success "Latest version: ${VERSION_TAG}"
    fi

    # Strip leading 'v' for archive naming (goreleaser drops it).
    VERSION_NUMBER="${VERSION_TAG#v}"
}

# ============================================================================
# Download + verify
# ============================================================================

download_and_install() {
    local archive_name download_url checksums_url tmpdir
    archive_name="${BINARY_NAME}_${VERSION_NUMBER}_${OS}_${ARCH}.tar.gz"
    download_url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${VERSION_TAG}/${archive_name}"
    checksums_url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${VERSION_TAG}/checksums.txt"

    tmpdir="$(mktemp -d)"
    # shellcheck disable=SC2064
    trap "rm -rf '$tmpdir'" EXIT

    info "Downloading ${archive_name}..."
    if ! curl -sfL -o "${tmpdir}/${archive_name}" "$download_url"; then
        fatal "Failed to download:\n  ${download_url}\n\nDoes a release exist for ${VERSION_TAG} on ${OS}/${ARCH}?"
    fi

    # Sanity: reject suspiciously small files (404 HTML, etc.)
    local file_size
    file_size="$(wc -c < "${tmpdir}/${archive_name}" | tr -d '[:space:]')"
    if [ "$file_size" -lt 1000 ]; then
        fatal "Downloaded file is suspiciously small (${file_size} bytes). Archive may not exist for this platform."
    fi
    success "Downloaded ${archive_name} (${file_size} bytes)"

    # Verify checksum — fail closed.
    info "Verifying checksum..."
    if ! curl -sfL -o "${tmpdir}/checksums.txt" "$checksums_url"; then
        fatal "Could not download checksums.txt from:\n  ${checksums_url}"
    fi

    local expected_checksum
    expected_checksum="$(grep " ${archive_name}\$" "${tmpdir}/checksums.txt" 2>/dev/null | awk '{print $1}' || true)"
    if [ -z "$expected_checksum" ]; then
        # Fallback: match without anchor (handles slight format variations)
        expected_checksum="$(grep "${archive_name}" "${tmpdir}/checksums.txt" 2>/dev/null | awk '{print $1}' | head -n 1 || true)"
    fi

    if [ -z "$expected_checksum" ]; then
        fatal "Archive '${archive_name}' not found in checksums.txt. Refusing to install unverified binary."
    fi

    local actual_checksum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_checksum="$(sha256sum "${tmpdir}/${archive_name}" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual_checksum="$(shasum -a 256 "${tmpdir}/${archive_name}" | awk '{print $1}')"
    else
        fatal "No sha256sum or shasum tool found. Cannot verify checksum."
    fi

    if [ "$actual_checksum" != "$expected_checksum" ]; then
        fatal "Checksum mismatch!\n  Expected: ${expected_checksum}\n  Got:      ${actual_checksum}"
    fi
    success "Checksum verified"

    # Extract
    info "Extracting ${BINARY_NAME}..."
    tar -xzf "${tmpdir}/${archive_name}" -C "$tmpdir" \
        || fatal "Failed to extract archive"

    if [ ! -f "${tmpdir}/${BINARY_NAME}" ]; then
        fatal "Binary '${BINARY_NAME}' not found in archive"
    fi

    # Install destination
    local install_dir="${DUCK_AI_INSTALL_DIR:-${HOME}/.local/bin}"
    mkdir -p "$install_dir"

    info "Installing to ${install_dir}/${BINARY_NAME}..."
    if ! cp "${tmpdir}/${BINARY_NAME}" "${install_dir}/${BINARY_NAME}" 2>/dev/null; then
        if command -v sudo >/dev/null 2>&1; then
            warn "Permission denied. Retrying with sudo..."
            sudo cp "${tmpdir}/${BINARY_NAME}" "${install_dir}/${BINARY_NAME}"
            sudo chmod +x "${install_dir}/${BINARY_NAME}"
        else
            fatal "Cannot write to ${install_dir}. Set DUCK_AI_INSTALL_DIR to a writable directory."
        fi
    else
        chmod +x "${install_dir}/${BINARY_NAME}"
    fi

    INSTALL_DIR="$install_dir"
    success "Installed ${BINARY_NAME} to ${install_dir}/${BINARY_NAME}"
}

# ============================================================================
# Next steps
# ============================================================================

print_next_steps() {
    echo ""
    printf '%b%bInstallation complete!%b\n' "$GREEN" "$BOLD" "$NC"
    echo ""

    # Warn if install dir is not in PATH
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *)
            warn "${INSTALL_DIR} is not in your PATH."
            printf '  Add this to your shell profile (~/.bashrc, ~/.zshrc, ...):\n'
            printf '    %bexport PATH="$PATH:%s"%b\n\n' "$DIM" "$INSTALL_DIR" "$NC"
            ;;
    esac

    printf '%bNext steps:%b\n' "$BOLD" "$NC"
    printf '  %b1.%b  %bduck-ai update%b   install Claude/Codex/OpenCode skills + commands\n' "$CYAN" "$NC" "$BOLD" "$NC"
    printf '  %b2.%b  %bduck-ai doctor%b   verify installation\n' "$CYAN" "$NC" "$BOLD" "$NC"
    printf '  %b3.%b  %bduck-ai%b          launch interactive TUI\n' "$CYAN" "$NC" "$BOLD" "$NC"
    echo ""
    printf '%bDocs: https://github.com/%s/%s%b\n' "$DIM" "$GITHUB_OWNER" "$GITHUB_REPO" "$NC"
    echo ""
}

# ============================================================================
# Main
# ============================================================================

main() {
    setup_colors

    step "Detecting platform"
    detect_platform
    check_prerequisites

    step "Resolving version"
    resolve_version

    step "Installing ${BINARY_NAME} ${VERSION_TAG}"
    download_and_install

    print_next_steps
}

main "$@"
