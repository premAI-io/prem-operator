# Prem Operator

Deploy AI models to Kubernetes using LocalAI, vLLM, DeepSpeed-MII, NVIDIA Triton and custom Docker images.

## Description

The Prem Operator provides a set of Kubernetes controllers and custom resource definitions (CRDs) which
deploy and manage AI models in a unified way. The major components are

- Deployment controller and resource definitions
- Model controller and resource definitions

The operator does not abstract away all the details between deploying on say vLLM or Triton. There are different
engines with different strengths. Instead it provides a common configuration where there is common ground.

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your operator will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

## Requirements

The Operator can be run without the following, but some features will be absent.

- An ingress controller (e.g. traefik)
- Helm
- The NVIDIA operator

**Note**: Helm is used to install Treafik when automatically creating a cluster with KIND (i.e. with `make kind-setup`).

## Deploying
### Helm

The Helm chart is available on Docker Hub it can be installed in the usual way.

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

### From source

1. Install Instances of Custom Resources:

```sh
make install
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/prem-operator:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/prem-operator:tag
```

### Example(locally with kind):

```
make kind-setup install deploy
kubectl apply -f example/test.yaml
curl http://foo.127.0.0.1.nip.io:8080/v1/completions -H "Content-Type: application/json" -d '{
     "model": "tinyllama-chat",
     "prompt": "Do you always repeat yourself?",
     "temperature": 0.1,
     "max_tokens": 50
   }'
```

There are more example configurations in `example/`.

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing

Feel free to open a pull request or create an issue. Often it makes sense to discuss new features before submitting a PR. This avoids
collisions and wasted effort, although it doesn't guarantee your code will be accepted.

We welcome draft PRs and experimental work. If you are going to code first then ask questions later, it's at least preferable to submit
early and often.

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

### Tutorial

1. [Run Mistral on K3s with vLLM from prem-operator source code](./docs/vllm.MD)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

