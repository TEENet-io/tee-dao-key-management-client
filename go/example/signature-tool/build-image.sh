#!/bin/bash

# TEENet Signature Tool - Docker Image Build and Compression Script
# This script builds the Docker image and compresses it for distribution

set -e

# Configuration
IMAGE_NAME="teenet-signature-tool"
IMAGE_TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"
OUTPUT_FILE="${IMAGE_NAME}-${IMAGE_TAG}.tar.gz"

echo "🔨 Building TEENet Signature Tool Docker image..."
echo "Image: ${FULL_IMAGE_NAME}"
echo "Output: ${OUTPUT_FILE}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Try to pull base images first (with timeout)
echo "📥 Pre-pulling base images..."
timeout 60 docker pull golang:1.24-alpine || echo "⚠️  Failed to pull golang image, will try during build"
timeout 60 docker pull alpine:latest || echo "⚠️  Failed to pull alpine image, will try during build"

# Build the Docker image with increased timeout
echo "📦 Building Docker image..."
DOCKER_BUILDKIT=1 docker build --network=host --progress=plain -t ${FULL_IMAGE_NAME} . 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Docker image built successfully!"
else
    echo "❌ Failed to build Docker image"
    echo ""
    echo "💡 Network issues detected. Try:"
    echo "   1. Check internet connection"
    echo "   2. Use local build instead: ./build-local.sh"
    echo "   3. Configure Docker proxy if behind firewall"
    exit 1
fi

# Save and compress the Docker image
echo "💾 Saving and compressing Docker image..."
docker save ${FULL_IMAGE_NAME} | gzip > ${OUTPUT_FILE}

if [ $? -eq 0 ]; then
    echo "✅ Docker image saved and compressed successfully!"
    echo "📁 Output file: ${OUTPUT_FILE}"
    
    # Display file size
    FILE_SIZE=$(du -h ${OUTPUT_FILE} | cut -f1)
    echo "📏 File size: ${FILE_SIZE}"
else
    echo "❌ Failed to save and compress Docker image"
    exit 1
fi

echo ""
echo "🚀 To load the image on another system:"
echo "   docker load < ${OUTPUT_FILE}"
echo ""
echo "🔧 To run the container:"
echo "   docker run -p 8080:8080 -e TEE_CONFIG_ADDR=localhost:50052 ${FULL_IMAGE_NAME}"