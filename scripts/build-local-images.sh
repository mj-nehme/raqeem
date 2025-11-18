#!/usr/bin/env bash
set -e

echo "🔨 Building local Docker images for Raqeem..."
echo ""

# Build devices backend
echo "📦 Building devices-backend..."
docker build -t raqeem/devices-backend:latest \
             -t raqeem/devices-backend:v0.2.0 \
             -f devices/backend/Dockerfile \
             devices/backend/

# Build mentor backend
echo "📦 Building mentor-backend..."
docker build -t raqeem/mentor-backend:latest \
             -t raqeem/mentor-backend:v0.2.0 \
             -f mentor/backend/Dockerfile \
             mentor/backend/

echo ""
echo "✅ Local images built successfully:"
echo "  • raqeem/devices-backend:latest"
echo "  • raqeem/devices-backend:v0.2.0"
echo "  • raqeem/mentor-backend:latest"
echo "  • raqeem/mentor-backend:v0.2.0"
echo ""
echo "💡 These images will be used by Kubernetes with pullPolicy: Never"
