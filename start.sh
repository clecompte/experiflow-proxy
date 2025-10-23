#!/bin/bash

# ExperiFlow Proxy - Local Development Startup Script

set -e

echo "🚀 Starting ExperiFlow Proxy..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ from https://golang.org/dl/"
    exit 1
fi

# Build the proxy
echo "📦 Building proxy..."
go build -o proxy ./cmd/proxy

# Set default environment variables
export PORT=${PORT:-8090}
export ORIGIN_URL=${ORIGIN_URL:-http://localhost:8080}
export EXPERIFLOW_API_URL=${EXPERIFLOW_API_URL:-http://localhost:8000}
export EXPERIMENT_IDS=${EXPERIMENT_IDS:-54ce9030-4da3-4866-8b25-6d956207f325}
export FAIL_OPEN=${FAIL_OPEN:-true}
export ENABLE_LOGGING=${ENABLE_LOGGING:-true}

echo "
✅ Configuration:
   Port:        $PORT
   Origin:      $ORIGIN_URL
   API:         $EXPERIFLOW_API_URL
   Experiments: $EXPERIMENT_IDS
"

# Run the proxy
echo "🎯 Starting proxy on http://localhost:$PORT"
echo "📝 Press Ctrl+C to stop"
echo ""

./proxy
