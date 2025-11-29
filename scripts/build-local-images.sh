#!/usr/bin/env bash
set -e

echo "🔨 Building local Docker images for Raqeem..."
echo ""

# Build devices backend
VERSION_FILE=$(dirname "$0")/../VERSION
VERSION=$(cat "$VERSION_FILE" | tr -d '\n')

echo "📦 Building devices-backend (version $VERSION)..."
docker build -t raqeem/devices-backend:latest \
             -t raqeem/devices-backend:${VERSION} \
             -f devices/backend/Dockerfile \
             devices/backend/

# Build mentor backend
echo "📦 Building mentor-backend (version $VERSION)..."
docker build -t raqeem/mentor-backend:latest \
             -t raqeem/mentor-backend:${VERSION} \
             -f mentor/backend/Dockerfile \
             mentor/backend/

echo ""
echo "✅ Local images built successfully:"
echo "  • raqeem/devices-backend:latest"
echo "  • raqeem/devices-backend:${VERSION}"
echo "  • raqeem/mentor-backend:latest"
echo "  • raqeem/mentor-backend:${VERSION}"
echo ""
echo "💡 These images will be used by Kubernetes with pullPolicy: Never"
