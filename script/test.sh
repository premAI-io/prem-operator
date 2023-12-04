#!/bin/bash -eu

CREATE_ONLY=${CREATE_ONLY:-false}
KUBE_VERSION=${KUBE_VERSION:-v1.22.7}
CLUSTER_NAME="${CLUSTER_NAME:-saas-controller-e2e}"

if ! kind get clusters | grep "$CLUSTER_NAME"; then
cat << EOF > kind.config
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    image: kindest/node:$KUBE_VERSION
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
    - containerPort: 32080
      hostPort: 8080
      protocol: TCP
    - containerPort: 32443
      hostPort: 8443
      protocol: TCP
    - containerPort: 32090
      hostPort: 9000
      protocol: TCP
EOF
cat << EOF > traefik-values.yaml
---
providers:
  kubernetesCRD:
    namespaces:
      - default
      - traefik
  kubernetesIngress:
    namespaces:
      - default
      - traefik

ports:
  traefik:
    expose: true
    nodePort: 32090
  web:
    nodePort: 32080
  websecure:
    nodePort: 32443
EOF
    kind create cluster --name $CLUSTER_NAME --config kind.config
    rm -rf kind.config
    helm repo add traefik https://traefik.github.io/charts
    helm repo update
    helm install traefik traefik/traefik --values traefik-values.yaml
    rm -rf traefik-values.yaml
fi

# IF CREATE_ONLY exits here
if [ "$CREATE_ONLY" = "true" ]; then
    exit 0
fi

set -e

kubectl cluster-info --context kind-$CLUSTER_NAME
echo "Sleep to give times to node to populate with all info"
kubectl wait --for=condition=Ready node/$CLUSTER_NAME-control-plane
export EXTERNAL_IP=$(kubectl get nodes -o jsonpath='{.items[].status.addresses[?(@.type == "InternalIP")].address}')
export BRIDGE_IP="172.18.0.1"
kubectl get nodes -o wide
cd $ROOT_DIR/tests &&  go run github.com/onsi/ginkgo/v2/ginkgo -r -v ./e2e
