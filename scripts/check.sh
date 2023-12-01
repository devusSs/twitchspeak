#!/bin/bash

# Display information message
echo "Checking for potential errors, unused vars, etc."
echo "This might take a moment, please wait"

# Run various checks
go mod tidy
go fmt ./...
go vet ./...
goimports -l ./
golangci-lint run ./...
gocritic check ./...
golines -w ./

# Display completion message
echo "Done, no errors detected"