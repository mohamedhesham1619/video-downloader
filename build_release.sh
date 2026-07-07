#!/bin/bash

# Exit on error
set -e

# Check if version is provided
if [ -z "$1" ]; then
    echo "Error: Version argument is required."
    echo "Usage: ./build_release.sh <version>"
    echo "Example: ./build_release.sh v1.1.0"
    exit 1
fi

VERSION=$1

echo "Building releases for version ${VERSION}..."

APP_NAME="downloader"
MAIN_FILE="./cmd/main.go"
RELEASE_DIR="release"

# Create release directory
mkdir -p "$RELEASE_DIR"

# Clean previous releases (except .gitkeep if it exists)
find "$RELEASE_DIR" -type f -not -name '.gitkeep' -delete

# Function to build and package
build_and_package() {
    local os=$1
    local arch=$2
    local extension=$3
    
    local os_arch_name="${APP_NAME}-${VERSION}-${os}-${arch}"
    local bin_name="${APP_NAME}"
    if [ "$os" == "windows" ]; then
        bin_name="${APP_NAME}.exe"
    fi

    echo "Building for $os/$arch..."
    
    # Create temporary directory for packaging
    local temp_dir="${RELEASE_DIR}/${os_arch_name}"
    mkdir -p "${temp_dir}"
    
    # Compile
    GOOS=$os GOARCH=$arch go build -o "${temp_dir}/${bin_name}" $MAIN_FILE
    
    # Create empty urls.txt
    touch "${temp_dir}/urls.txt"
    
    # Copy the instructions file
    cp HOW_TO_USE.txt "${temp_dir}/"
    
    # Package
    echo "Packaging ${os_arch_name}..."
    cd "${RELEASE_DIR}"
    
    if [ "$extension" == "tar.gz" ]; then
        tar -czf "${os_arch_name}.tar.gz" "${os_arch_name}/"
    elif [ "$extension" == "zip" ]; then
        zip -r -q "${os_arch_name}.zip" "${os_arch_name}/"
    fi
    
    # Clean up temp dir
    rm -rf "${os_arch_name}"
    cd ..
}

# Build Linux
build_and_package "linux" "amd64" "tar.gz"

# Build macOS (Intel & Apple Silicon)
build_and_package "darwin" "amd64" "tar.gz"
build_and_package "darwin" "arm64" "tar.gz"

# Build Windows
build_and_package "windows" "amd64" "zip"

echo "Releases built successfully in ${RELEASE_DIR}/"
