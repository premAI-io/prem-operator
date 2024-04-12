# FAQ

## Why use kubernetes at all?

Using kubernetes is a great way to deploy AI models for enterprise solutions where scalability and fault tolerance are critical.
But it's also a great way to deploy AI models for non-enterprise solutions where maximizing resource usage and customization are more important.

### Pros

1. **Easy deployment and scaling**. With the kubernetes, you can easily deploy models as "deployment" and scale them up with properly load balanced "replicas".
2. **Easy management**. With the kubernetes, you can easily manage AI models across multiple engines with well-defined deterministic deployment strategies and configurations.

### Cons

1. **Complexity of initial setup and maintainence**. With the kubernetes, it's quite complex to setup.
2. **Fixing and debugging**. With the kubernetes, it's quite complex to fix issues and debug them as services.

## How do I manually upgrade the operator?

1. To upgrade the operator, simply run `helm upgrade prem-operator --recreate-pods`.
2. `--recreate-pods` is required if you want updates to take effect without destroying to the pods.
3. In case of issues you can use `helm rollback prem-operator ~1` to rollback to previous release.

## What are the key features of the operator?

1. **Unified deployment framework**. It abstracts and simplifies the deployment process of AI models on Kubernetes, making it quicker and more efficient.
2. **Engine versatility**. While it simplified common configuration aspects, and offers support for multiple engines, simplifying the deployment strategies for each engine.
3. **Flexible architecture**. It offers a consistent and flexible architecture, allowing the operator to freely work with most other kubernetes tools and resources.
4. **Dashboard**. It provides an easy-to-use dashboard that allows you to monitor and manage AI models in real-time.

## Can the operator be used outside of Kubernetes?

No, the operator is not designed to be used outside of Kubernetes, and depends on Kubernetes Operator SDK to interact with the cluster.

## How can I contribute to the operator?

Please refer to the [CONTRIBUTING.md](docs/contributing.md) file for details.

## How can I report an issue?

Please refer to the [ISSUES.md](docs/issues.md) file for details.
