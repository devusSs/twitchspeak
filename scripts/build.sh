#!/bin/bash

# Detect build OS
BUILD_OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Detect build architecture
BUILD_ARCH=$(uname -m)

# Set the build version / check if it is set
[ "${TS_BUILD_VERSION}" ] || { echo "TS_BUILD_VERSION is not set"; exit 1; }

# Display compilation information
echo "Compiling app for ${BUILD_OS} (${BUILD_ARCH})..."

# Run go mod tidy
go mod tidy

# Build the application
CGO_ENABLED=0 GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH}" go build \
  -v -trimpath \
  -ldflags="-s -w -X 'main.buildVersion=${TS_BUILD_VERSION}' -X 'main.buildDate=$(date)' -X 'main.buildGitCommit=$(git rev-parse HEAD)'" \
  -o "./.release/twitchspeak_${BUILD_OS}_${BUILD_ARCH}/" ./...

# Display completion message
echo "Done building app"