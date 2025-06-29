#!/bin/bash

# Version bump script for go-envsync
# Usage: ./scripts/version-bump.sh [patch|minor|major]

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
readonly VERSION_FILE="$PROJECT_ROOT/.release-version"

# Constants for version bumping
readonly PATCH_INCREMENT=1
readonly MINOR_INCREMENT=1
readonly MAJOR_INCREMENT=1

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage information
show_usage() {
    cat << 'EOF'
Usage: $0 [patch|minor|major]

Version Bump Script for go-envsync

Options:
  patch   Increment patch version (X.Y.Z -> X.Y.Z+1)
  minor   Increment minor version (X.Y.Z -> X.Y+1.0)
  major   Increment major version (X.Y.Z -> X+1.0.0)

Examples:
  $0 patch    # 0.1.0 -> 0.1.1
  $0 minor    # 0.1.1 -> 0.2.0
  $0 major    # 0.2.0 -> 1.0.0
EOF
}

# Parse semantic version
parse_version() {
    local version="$1"
    
    if [[ ! "$version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        print_error "Invalid version format: $version"
        exit 1
    fi
    
    MAJOR_VERSION="${BASH_REMATCH[1]}"
    MINOR_VERSION="${BASH_REMATCH[2]}"
    PATCH_VERSION="${BASH_REMATCH[3]}"
}

# Bump version based on type
bump_version() {
    local bump_type="$1"
    local current_version="$2"
    
    parse_version "$current_version"
    
    case "$bump_type" in
        patch)
            PATCH_VERSION=$((PATCH_VERSION + PATCH_INCREMENT))
            ;;
        minor)
            MINOR_VERSION=$((MINOR_VERSION + MINOR_INCREMENT))
            PATCH_VERSION=0
            ;;
        major)
            MAJOR_VERSION=$((MAJOR_VERSION + MAJOR_INCREMENT))
            MINOR_VERSION=0
            PATCH_VERSION=0
            ;;
        *)
            print_error "Invalid bump type: $bump_type"
            show_usage
            exit 1
            ;;
    esac
    
    echo "${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}"
}

# Update version file
update_version_file() {
    local new_version="$1"
    
    echo "$new_version" > "$VERSION_FILE"
    
    if [[ $? -eq 0 ]]; then
        print_info "Updated version file: $VERSION_FILE"
    else
        print_error "Failed to update version file"
        exit 1
    fi
}

# Validate git repository state
validate_git_state() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
    
    if [[ -n "$(git status --porcelain)" ]]; then
        print_warn "Working directory has uncommitted changes"
        print_warn "Consider committing changes before version bump"
    fi
}

# Create git tag
create_git_tag() {
    local version="$1"
    local tag_name="v$version"
    
    if git tag -a "$tag_name" -m "Release $tag_name"; then
        print_info "Created git tag: $tag_name"
    else
        print_error "Failed to create git tag"
        exit 1
    fi
}

# Main function
main() {
    local bump_type="${1:-}"
    
    if [[ -z "$bump_type" ]]; then
        print_error "Missing required argument"
        show_usage
        exit 1
    fi
    
    if [[ "$bump_type" == "--help" || "$bump_type" == "-h" ]]; then
        show_usage
        exit 0
    fi
    
    # Validate git state
    validate_git_state
    
    # Read current version
    if [[ ! -f "$VERSION_FILE" ]]; then
        print_error "Version file not found: $VERSION_FILE"
        exit 1
    fi
    
    local current_version
    current_version="$(cat "$VERSION_FILE")"
    
    print_info "Current version: $current_version"
    
    # Calculate new version
    local new_version
    new_version="$(bump_version "$bump_type" "$current_version")"
    
    print_info "New version: $new_version"
    
    # Confirm with user
    echo -n "Proceed with version bump? [y/N] "
    read -r response
    
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_info "Version bump cancelled"
        exit 0
    fi
    
    # Update version file
    update_version_file "$new_version"
    
    # Create git commit
    if git add "$VERSION_FILE" && git commit -m "bump version to $new_version"; then
        print_info "Created version bump commit"
    else
        print_error "Failed to create git commit"
        exit 1
    fi
    
    # Create git tag
    create_git_tag "$new_version"
    
    print_info "Version bump completed successfully!"
    print_info "Next steps:"
    print_info "  1. Push changes: git push origin main"
    print_info "  2. Push tags: git push origin --tags"
    print_info "  3. Create GitHub release"
}

# Execute main function
main "$@" 