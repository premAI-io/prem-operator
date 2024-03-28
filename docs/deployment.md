# Deployment

## Requirements

- K8s cluster
- Helm
- Ingress controller (e.g. Traefik)
- Nvidia GPU Operator

## Artifacts
Prem-Operator has three artifacts:
1. [Source Code](https://github.com/premAI-io/saas-controller)
1. [Prem-Operator Docker image](#container-images)
1. [Prem-Operator Helm chart](https://hub.docker.com/r/premai/prem-operator-chart)

## Installation
After setting up K8s cluster, install [Nvidia GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/getting-started.html) and [Traefik](https://doc.traefik.io/traefik/getting-started/install-traefik/#use-the-helm-chart) as an ingress controller.
Note that Nvidia GPU Operator is required for GPU support and Traefik is required for handling ingress traffic.
Now install Prem-Operator using Helm:
```bash
$ helm install <my-release> oci://registry-1.docker.io/premai/prem-operator-chart
```

### Flux

If you are using Flux and GitOps then you can commit something like the below and include it
in a Kustomization manifest.

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: ai
  namespace: prem-operator
spec:
  interval: 1m
  chart:
    spec:
      chart: prem-operator-chart
      sourceRef:
        kind: HelmRepository
        name: prem-operator
        namespace: prem-operator
      version: "x.x.x" 
      interval: 1m
  install:
    crds: CreateReplace
  upgrade:
    crds: CreateReplace
```

`x.x.x` should be replaced with a real version number.
