## 📡 Prem Operator: On premise AI with Kubernetes

### 🗺 Overview

Deploy AI models and apps to Kubernetes without hitting every pitfall.

### 🚀 Quick Start

Deploy the operator and chat with Hermes in 2 or 3 steps. Note that you need a NVIDIA GPU with 8GB of RAM.

1. Install the NVIDIA Operator
2. Install the Prem Operator
    ```
    $ helm install latest oci://registry-1.docker.io/premai/prem-operator-chart
    ```

3. Deploy the Hermes + Big-AGI example and forward the ports
    ```
    $ kubectl apply -f examples/big-agi.yaml
    $ kubectl port-forward services/big-agi-service 3000:3000
    ```

If you browse to localhost:3000 you'll be able to begin chatting as soon as
LocalAI downloads the model.

https://github.com/richiejp/prem-operator/assets/988098/0f06b254-a1a0-4ae5-815a-ed84998f5c89

### 🔗 Useful Links

- **Guides**
    - [🦜️🔗**Langchain**](./docs/guides/langchain.md)
    - [🧩**Ingress**](./docs/guides/ingress.md)
    - [🌐**Managed Clusters (GCP, AWS)**](./docs/guides/managed_cluster.md)
    - [📜**Getting started**](./docs/getting_started.md)
    - [📦**Deployment**](./docs/deployment.md)
    - [👩‍💻**Developer**](./docs/developer_guide.md)
    - [🧪**vLLM**](./docs/vllm.md)
- **Topics**
    - [🔍**Frequently Asked Questions**](./docs/faq.md)
    - [🤝**Contributing**](./docs/contributing.md)
    - [🎗️**Issues**](./docs/issues.md)
    - [🔧**Troubleshooting**](./docs/troubleshooting.md)

### Simplified Architecture Diagram 

![simpligied architecture diagram prem-operator](https://github.com/premAI-io/prem-operator/assets/19930870/0b800db2-2c04-4b69-9672-d9f3c982507d)

### 💌 Contact & Getting Help

We'd love to hear from you! Feel free to create an issue or reach out to us on [Prem's Discord](https://discord.com/invite/kpKk6vYVAn) on the operator channel.

### 📝 License

Operator SDK is under Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.

