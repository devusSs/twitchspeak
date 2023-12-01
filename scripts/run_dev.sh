#!/bin/bash

# Detect build OS
BUILD_OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Detect build architecture
BUILD_ARCH=$(uname -m)

# Create testing directory
mkdir -p ./.testing 

# Copy the built binary to the testing directory
cp ./.release/twitchspeak_"${BUILD_OS}"_"${BUILD_ARCH}"/twitchspeak ./.testing/twitchspeak

# Run the application with specific parameters
./.testing/twitchspeak -l .logs_dev -c .env --console --debug