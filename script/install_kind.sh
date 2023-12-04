#!/bin/bash

# Update the package index
sudo apt-get update

# Install prerequisites
sudo apt-get install -y curl

# Install Kind
curl -Lo ./kind "https://kind.sigs.k8s.io/dl/v0.11.1/kind-$(uname)-amd64"
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Verify Kind installation
kind --version

echo "Kind and k9s installation complete."

