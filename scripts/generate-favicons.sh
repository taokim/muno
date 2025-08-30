#!/bin/bash

# Script to generate favicon files from the MUNO logo
# Requires ImageMagick (brew install imagemagick)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ASSETS_DIR="$PROJECT_ROOT/assets"
SOURCE_IMAGE="$ASSETS_DIR/muno-logo.png"

echo "🐙 MUNO Favicon Generator"
echo "========================"
echo ""

# Check if ImageMagick is installed
if ! command -v convert &> /dev/null; then
    echo "❌ ImageMagick is not installed."
    echo "Please install it first:"
    echo "  macOS: brew install imagemagick"
    echo "  Ubuntu: sudo apt-get install imagemagick"
    exit 1
fi

# Check if source image exists
if [ ! -f "$SOURCE_IMAGE" ]; then
    echo "❌ Source image not found: $SOURCE_IMAGE"
    echo "Please ensure muno-logo.png exists in the assets directory."
    exit 1
fi

echo "📝 Generating favicons from $SOURCE_IMAGE..."
echo ""

# Create different sizes for various platforms
echo "Creating favicon-16x16.png..."
convert "$SOURCE_IMAGE" -resize 16x16 "$ASSETS_DIR/favicon-16x16.png"

echo "Creating favicon-32x32.png..."
convert "$SOURCE_IMAGE" -resize 32x32 "$ASSETS_DIR/favicon-32x32.png"

echo "Creating favicon-48x48.png..."
convert "$SOURCE_IMAGE" -resize 48x48 "$ASSETS_DIR/favicon-48x48.png"

echo "Creating favicon-64x64.png..."
convert "$SOURCE_IMAGE" -resize 64x64 "$ASSETS_DIR/favicon-64x64.png"

echo "Creating favicon-96x96.png..."
convert "$SOURCE_IMAGE" -resize 96x96 "$ASSETS_DIR/favicon-96x96.png"

echo "Creating favicon-128x128.png..."
convert "$SOURCE_IMAGE" -resize 128x128 "$ASSETS_DIR/favicon-128x128.png"

echo "Creating apple-touch-icon.png (180x180)..."
convert "$SOURCE_IMAGE" -resize 180x180 "$ASSETS_DIR/apple-touch-icon.png"

echo "Creating android-chrome-192x192.png..."
convert "$SOURCE_IMAGE" -resize 192x192 "$ASSETS_DIR/android-chrome-192x192.png"

echo "Creating android-chrome-512x512.png..."
convert "$SOURCE_IMAGE" -resize 512x512 "$ASSETS_DIR/android-chrome-512x512.png"

echo "Creating mstile-150x150.png for Windows..."
convert "$SOURCE_IMAGE" -resize 150x150 "$ASSETS_DIR/mstile-150x150.png"

# Create favicon.ico with multiple sizes
echo "Creating favicon.ico (multi-resolution)..."
convert "$SOURCE_IMAGE" -resize 16x16 -resize 32x32 -resize 48x48 -resize 64x64 "$ASSETS_DIR/favicon.ico"

# Create social media preview image (1200x630 for Open Graph)
echo "Creating social media preview image..."
# Create a canvas with the octopus logo and text
convert -size 1200x630 xc:'#4a90e2' \
    \( "$SOURCE_IMAGE" -resize 300x300 \) -geometry +450+100 -composite \
    -font Helvetica-Bold -pointsize 72 -fill white \
    -annotate +0+480 "MUNO" -gravity center \
    -font Helvetica -pointsize 36 -fill white \
    -annotate +0+550 "Multi-repository UNified Orchestration" -gravity center \
    "$ASSETS_DIR/muno-social-preview.png"

# Create Twitter card image (1200x675 for better Twitter display)
echo "Creating Twitter card image..."
convert -size 1200x675 xc:'#4a90e2' \
    \( "$SOURCE_IMAGE" -resize 320x320 \) -geometry +440+100 -composite \
    -font Helvetica-Bold -pointsize 80 -fill white \
    -annotate +0+500 "MUNO" -gravity center \
    -font Helvetica -pointsize 40 -fill white \
    -annotate +0+580 "Multi-repository UNified Orchestration" -gravity center \
    "$ASSETS_DIR/muno-twitter-card.png"

# Create safari-pinned-tab.svg (simplified monochrome version)
echo "Creating Safari pinned tab icon..."
# Note: This creates a simple PNG version. For a true SVG, you'd need to trace the image
convert "$SOURCE_IMAGE" -resize 512x512 -colorspace Gray "$ASSETS_DIR/safari-pinned-tab.png"

echo ""
echo "✅ Favicon generation complete!"
echo ""
echo "📁 Generated files in $ASSETS_DIR:"
ls -la "$ASSETS_DIR"/*.png "$ASSETS_DIR"/*.ico 2>/dev/null | awk '{print "  - "$NF}'
echo ""
echo "📌 Next steps:"
echo "  1. Commit these files to your repository"
echo "  2. The files will be used for:"
echo "     - Browser tabs (favicon.ico)"
echo "     - Mobile bookmarks (apple-touch-icon.png)"
echo "     - Social media sharing (muno-social-preview.png)"
echo "     - PWA installation (android-chrome-*.png)"
echo ""
echo "🐙 Your MUNO branding is ready for the web!"