# Install k9s
curl -sS https://webinstall.dev/k9s | bash
sudo mv $HOME/.local/bin/k9s /usr/local/bin/k9s

# Verify k9s installation
k9s version
