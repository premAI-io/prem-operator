curl -sfL https://get.k3s.io | sh -
sudo ln -s /usr/local/bin/k3s /usr/bin/k3s
export KUBECONFIG=~/.kube/config
mkdir ~/.kube 2> /dev/null
sudo k3s kubectl config view --raw > "$KUBECONFIG"
chmod 600 "$KUBECONFIG"
echo 'export KUBECONFIG=~/.kube/config' >> ~/.bashrc
echo 'export KUBECONFIG="~/.kube/config' >> ~/.profile
source .bashrc