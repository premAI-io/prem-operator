# Getting Started

## Prerequisites

1. Install [Helm](https://helm.sh) on your system.
2. Install [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) on your system.

## Setup the cluster

1. Setup local or managed cluster(feel free to use our guide at ./guides/managed_cluster.md) and get the kube config for the cluster, and store it at `~/.kube/config`.
2. Add the following line to your `~/.bashrc` or `~/.zshrc`: `export KUBECONFIG=~/.kube/config`. And source the file.
3. Run `kubectl cluster-info` to verify that the cluster is ready to use.
4. Make sure the cluster is reachable and setup any form of ingress you would like, e.g. Traefik or NGINX (our guide at ./guides/ingress.md).

## Deploying the Operator

Installing the operator is straightforward if you are open to using `helm`.

1. Add the helm repository with `helm repo add prem-operator-charts https://prem-operator.github.io/charts`.
2. Install the operator with `helm install prem-operator prem-operator-charts/prem-operator -n prem-operator --create-namespace`.
3. Verify that the operator is running with `kubectl get pods -n prem-operator`.

## Using the operator

### Model Maps

Model map resources are used by the controller for defining the configurations for models so that they can be reused by the deployments.

For example,

```yaml
apiVersion: premlabs.io/v1alpha1
kind: AIModelMap
metadata:
  name: mixtral-8x7b-instruct-v0.1
spec:
  localai:
    - variant: base
      uri: "https://huggingface.co/mudler/LocalAI/resolve/main/examples/configurations/mixtral/mixtral.yaml?download=true"
  vllm:
    - variant: base
      uri: "mistralai/Mixtral-8x7B-Instruct-v0.1"
```

You can specify different variants as well as different engine versions for each model, and even the data type and quantization.

```yaml
apiVersion: premlabs.io/v1alpha1
kind: AIModelMap
metadata:
  name: tinyllama-1.1b-chat-v0.1
spec:
  localai:
    - variant: q4-k-m
      uri: "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v0.3-GGUF/resolve/main/tinyllama-1.1b-chat-v0.3.Q4_K_M.gguf"
  vllm:
    - variant: awq
      uri: "TheBloke/TinyLlama-1.1B-Chat-v0.1-AWQ"
      dataType: "float16"
```

### AutoNodeLabeler

Auto Node labelling is used by the controller for defining the labels for specific nodes using generic match expressions, this is useful when you have to automatically add a large number of labels to your nodes.

```yaml
apiVersion: premlabs.io/v1alpha1
kind: AutoNodeLabeler
metadata:
  name: auto-node-labeler
spec:
  labels:
    foo/bar.com: baz
  matchExpression:
    key: kubernetes.io/arch
    operator: In
    values:
      - amd64
```

### AI Deployments

AI Deployments are at the core of the Prem Operator. They are used to deploy and manage AI models, and their associated resources.

```yaml
apiVersion: premlabs.io/v1alpha1
kind: AIDeployment 
metadata:
  name: simple
spec:
  engine:
    name: "localai" 
  endpoint:
    - domain: "tinyllama.localai.yourdomain.com"
      port: 8080 
  models:
    - uri: tinyllama-chat
```

## More  

Read the docs and the guides to learn more, about various ways you can use the prem-operator and our other projects.
