#!/bin/bash

# Update the package index
sudo apt-get update

# Install dependencies
sudo apt-get install -y curl apt-transport-https gnupg

# Add Helm's repository
curl -s https://baltocdn.com/helm/signing.asc | sudo apt-key add -
echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list

# Update the package index
sudo apt-get update

# Install Helm
sudo apt-get install helm

# Verify the installation
helm version

echo "Helm installation complete."

