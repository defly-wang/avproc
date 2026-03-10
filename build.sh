#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/version.env"

APP_NAME="avproc"
ARCH="amd64"
OUTPUT_DIR="dist"

build_linux() {
    echo "Building Linux..."
    mkdir -p "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/usr/bin"
    mkdir -p "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/usr/share/applications"
    mkdir -p "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/usr/share/icons/hicolor/256x256/apps"

    GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/usr/bin/$APP_NAME" .

    cp assets/icon.png "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/usr/share/icons/hicolor/256x256/apps/$APP_NAME.png" 2>/dev/null || true

    mkdir -p "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/DEBIAN"
    cat > "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/DEBIAN/control" << EOF
Package: $APP_NAME
Version: $VERSION
Section: video
Priority: optional
Depends: ffmpeg
Architecture: $ARCH
Maintainer: $AUTHOR
Description: A video processing tool based on FFmpeg
 A simple video processing tool with GUI, supporting convert, crop, merge, and more.
EOF

    cat > "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH/usr/share/applications/$APP_NAME.desktop" << EOF
[Desktop Entry]
Name=AVProc
Comment=Video processing tool
Exec=$APP_NAME
Icon=$APP_NAME
Terminal=false
Type=Application
Categories=AudioVideo;Video;
EOF

    dpkg-deb --build "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH" "$OUTPUT_DIR/${APP_NAME}_${VERSION}_${ARCH}.deb"
    rm -rf "$OUTPUT_DIR/$APP_NAME-$VERSION-linux-$ARCH"
    echo "Built: $OUTPUT_DIR/${APP_NAME}_${VERSION}_${ARCH}.deb"
}

build_windows() {
    echo "Building Windows..."
    mkdir -p "$OUTPUT_DIR/$APP_NAME-$VERSION-win-$ARCH"

    if [ -d "ffmpeg" ]; then
        mkdir -p "$OUTPUT_DIR/$APP_NAME-$VERSION-win-$ARCH/ffmpeg"
        cp ffmpeg/* "$OUTPUT_DIR/$APP_NAME-$VERSION-win-$ARCH/ffmpeg/" 2>/dev/null || true
    fi

    GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/$APP_NAME-$VERSION-win-$ARCH/$APP_NAME.exe" .

    cd "$OUTPUT_DIR/$APP_NAME-$VERSION-win-$ARCH"
    zip -r "../${APP_NAME}_${VERSION}_win-${ARCH}.zip" .
    cd - > /dev/null
    echo "Built: $OUTPUT_DIR/${APP_NAME}_${VERSION}_win-${ARCH}.zip"
}

clean() {
    rm -rf "$OUTPUT_DIR"
    echo "Cleaned $OUTPUT_DIR"
}

case "${1:-all}" in
    linux)
        mkdir -p "$OUTPUT_DIR"
        build_linux
        ;;
    windows)
        mkdir -p "$OUTPUT_DIR"
        build_windows
        ;;
    all)
        mkdir -p "$OUTPUT_DIR"
        build_linux
        build_windows
        ;;
    clean)
        clean
        ;;
    *)
        echo "Usage: $0 {linux|windows|all|clean}"
        exit 1
        ;;
esac
