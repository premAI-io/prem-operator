# Developer guide

## Project Overview

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster
The major components are:
1. **AI Deployment Custom Resource with Controller**:<br>
   AIDeployment is a custom Kubernetes resource that encapsulates the configuration necessary for deploying and managing AI models within a Kubernetes cluster. It allows users to specify details about the AI engine, model parameters, computational resource requirements, networking settings like endpoints, services, and ingresses, as well as environmental variables and arguments for model deployment. This resource aims to streamline the deployment process of AI models, making it easier to manage, scale, and update AI deployments in a cloud-native ecosystem.
2. **AIModelMap Custom Resource with Controller**:<br>
   The AIModelMap Custom Resource (CR) is a Kubernetes resource defined to facilitate the mapping and management of artificial intelligence (AI) model specifications across various execution engines, such as TensorRT, DeepSpeed-Mii, LocalAI, and VLLM. It allows for the specification of key details like the model's data type, engine configuration, quantization settings, and access URIs, alongside variant-specific configurations. By serving as a centralized repository for model specifications, AIModelMap enables consistent, efficient, and scalable deployment of AI models across multiple Kubernetes deployments, streamlining the management of model configurations and fostering reuse and flexibility in AI deployments within the Kubernetes ecosystem.
3. **AutoNodeLabeler with Controller**:<br>
   The AutoNodeLabeler is a Kubernetes custom resource (CR) designed to automatically apply labels to nodes based on specified criteria, such as hardware configurations. It enables precise and dynamic scheduling of workloads by labeling nodes with specific attributes like GPU types and sizes. This CR facilitates efficient resource utilization, cluster segmentation, and automation in node labeling, improving both cluster management and workload performance.

## Requirements

- Docker
- Kind
- Kubectl
- Kustomize
- Go
- Helm

## Development and Testing

The Prem-Operator needs to be developed and tested in a Kubernetes environment. For local testing, we recommend using [KIND](https://sigs.k8s.io/kind), which allows you to run a Kubernetes cluster within Docker containers. <br>

To facilitate the development process, we provide various Makefile targets leveraging KIND as the Kubernetes cluster. Run make --help to see all available make targets. <br>

Please note that certain components, like vLLM engines, require GPU support for execution. As KIND does not offer GPU support, alternatives like K3s or any Kubernetes cluster with GPU capabilities should be considered for these cases. For detailed instructions on running Mistral on K3s with vLLM from the Prem-Operator source code, refer to  [this guide](./vllm.md) for more information.

#### Installing Prem-Operator: Process Overview

Installing the Prem-Operator involves a series of steps designed to prepare your local development environment for both development and testing. Here’s what happens at each step:

- **Install KIND:** KIND (Kubernetes IN Docker) is a tool for running local Kubernetes clusters using Docker container “nodes.” KIND is particularly useful for development and testing of Kubernetes applications like Prem-Operator. The installation sets up KIND on your machine, enabling you to create a local cluster.

- **Create a KIND Cluster:** This step initializes a new Kubernetes cluster on your local machine using KIND. The cluster simulates a real Kubernetes environment, allowing you to test Prem-Operator in conditions resembling its intended runtime environment.

- **Install Traefik Ingress Controller:** Traefik is used as an ingress controller in Kubernetes clusters. It routes requests to correct services based on the incoming request’s host or path. Installing Traefik ensures that your local cluster can manage external access to its services, mimicking production ingress behavior.

- **Build Prem-Operator Docker Image and Load It to KIND Cluster:** Prem-Operator needs to run as a container within the Kubernetes cluster. This step compiles the Prem-Operator code into a Docker image and loads this image into your KIND cluster, making it available for deployment.

- **Install CRDs into the Cluster:** Custom Resource Definitions (CRDs) extend Kubernetes by defining new, custom resources. Prem-Operator relies on these custom resources to operate. Installing the CRDs updates your cluster to understand and manage the custom resources that Prem-Operator uses.

- **Deploy the Prem-Operator Controller:** Finally, this step deploys the Prem-Operator controller into your Kubernetes cluster. The controller is the core component that monitors for changes to resources and applies the necessary logic to react to these changes, effectively managing the Prem-Operator's operational logic within the cluster.


#### Install Kind 
To install KIND, go to [official page](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) and follow the instructions for your operating system, or run the following script for Ubuntu:
```bash
./../script/install_kind.sh
```

#### Create a Kind cluster with Traefik, build prem-operator Docker image and load it to Kind cluster
```bash
make kind-setup
```

#### Install CRDs into the cluster
```bash
make install
```

#### Deploy prem-operator controller
```bash
make deploy
```

#### Run test 
```bash
make test
```

#### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```


#### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

#### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Run AI Model inside Engine
Check [examples](./../examples) of AI Model deployment inside different Engines.

Bellow is an example of deploying tinyllama inside LocalAi engine:
```
kubectl apply -f example/test.yaml
```

After deploying you can infer the model using the following curl command:
```
curl http://foo.127.0.0.1.nip.io:8080/v1/completions -H "Content-Type: application/json" -d '{
"model": "tinyllama-chat",
"prompt": "Do you always repeat yourself?",
"temperature": 0.1,
"max_tokens": 50
}'
```
