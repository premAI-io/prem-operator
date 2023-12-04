#!/bin/bash

# Define Go version
GO_VERSION="1.20"

# Update the package index
sudo apt-get update

# Install wget if it's not already installed
sudo apt-get install -y wget

# Download the Go binary
wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz

# Extract the tarball
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

# Move Go binaries to /usr/local/bin
sudo ln -s /usr/local/go/bin/go /usr/local/bin/go

# Verify Go installation
go version

echo "Go installation complete."

