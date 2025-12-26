#!/bin/bash
set -e

# Extract version from control file if not provided
if [ -z "$1" ]; then
    VERSION=$(grep "^Version:" deb/control | awk '{print $2}')
    echo "Using version from control file: ${VERSION}"
else
    VERSION=$1
    echo "Using provided version: ${VERSION}"
fi

ARCH=${2:-amd64}
BUILD_DIR="build/deb"
PACKAGE_NAME="athens-multipat"
PACKAGE_DIR="${BUILD_DIR}/${PACKAGE_NAME}_${VERSION}_${ARCH}"

echo "Building ${PACKAGE_NAME} v${VERSION} for ${ARCH}..."

# Clean previous builds
rm -rf ${BUILD_DIR}

# Create package directory structure
mkdir -p ${PACKAGE_DIR}/DEBIAN
mkdir -p ${PACKAGE_DIR}/usr/local/bin
mkdir -p ${PACKAGE_DIR}/etc/athens
mkdir -p ${PACKAGE_DIR}/lib/systemd/system
mkdir -p ${PACKAGE_DIR}/var/lib/athens
mkdir -p ${PACKAGE_DIR}/var/log/athens

# Build binaries
echo "Building Athens proxy..."
CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build \
    -ldflags="-w -s -X github.com/gomods/athens/pkg/build.version=${VERSION} -X github.com/gomods/athens/pkg/build.buildDate=$(date -u +%Y-%m-%d-%H:%M:%S-%Z)" \
    -o ${PACKAGE_DIR}/usr/local/bin/athens \
    ./cmd/proxy

# Make binary executable
chmod +x ${PACKAGE_DIR}/usr/local/bin/athens

# Copy control files
cp deb/control ${PACKAGE_DIR}/DEBIAN/
sed -i "s/Version: .*/Version: ${VERSION}/" ${PACKAGE_DIR}/DEBIAN/control
sed -i "s/Architecture: .*/Architecture: ${ARCH}/" ${PACKAGE_DIR}/DEBIAN/control

cp deb/postinst ${PACKAGE_DIR}/DEBIAN/
cp deb/postrm ${PACKAGE_DIR}/DEBIAN/
chmod +x ${PACKAGE_DIR}/DEBIAN/postinst
chmod +x ${PACKAGE_DIR}/DEBIAN/postrm

# Copy configuration example and systemd service
# Note: We only copy the example file. The postinst script will create
# config.toml from the example if it doesn't exist.
cp deb/config.toml.example ${PACKAGE_DIR}/etc/athens/config.toml.example
cp deb/athens.service ${PACKAGE_DIR}/lib/systemd/system/

# Build the package
echo "Building debian package..."
dpkg-deb --build ${PACKAGE_DIR}

echo "Package built: ${BUILD_DIR}/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
ls -lh ${BUILD_DIR}/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb
