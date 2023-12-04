#!/bin/bash

# Update the package index
sudo apt-get update

# Install make and curl
sudo apt-get install -y make curl

# Install jq
sudo apt install jq

# Verify the installations
make --version
curl --version
jq --version 

echo "Installation of make, curl, jq complete."

