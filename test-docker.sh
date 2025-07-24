#!/bin/bash

set -e

echo "🚀 Starting Docker-based testing for Chess Engine..."

# Clean up any existing containers and images
echo "🧹 Cleaning up existing containers and images..."
docker stop chess-engine-test 2>/dev/null || true
docker rm chess-engine-test 2>/dev/null || true
docker rmi chess-engine:test 2>/dev/null || true

# Build the Docker image
echo "🔨 Building Docker image..."
docker build -t chess-engine:test .

# Test that the container starts successfully
echo "🧪 Testing container startup..."
docker run -d --name chess-engine-test -p 8080:8080 chess-engine:test

# Wait for the server to start
echo "⏳ Waiting for server to start..."
sleep 5

# Check if the server is responding
echo "🌐 Testing server health..."
if curl -f http://localhost:8080/ > /dev/null 2>&1; then
    echo "✅ Server is responding successfully!"
    
    # Test API endpoint
    echo "🔍 Testing API endpoint..."
    if curl -f http://localhost:8080/api/state > /dev/null 2>&1; then
        echo "✅ API endpoint is working!"
    else
        echo "❌ API endpoint test failed"
        exit 1
    fi
else
    echo "❌ Server health check failed"
    docker logs chess-engine-test
    exit 1
fi

# Show container logs for verification
echo "📋 Container logs:"
docker logs chess-engine-test | tail -10

# Clean up
echo "🧹 Cleaning up test container..."
docker stop chess-engine-test
docker rm chess-engine-test

echo "🎉 All Docker tests passed! The refactored code works correctly in Docker."
echo ""
echo "To run the application:"
echo "  docker run -d -p 8080:8080 chess-engine:test"
echo ""
echo "To access the application:"
echo "  http://localhost:8080" 