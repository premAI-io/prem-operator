# Troubleshooting

If you are new to Kubernetes then we would recommend using [K9s](https://k9scli.io/) in addition to kubectl.

In general whenever there is a problem.

1. What do the pod logs say?
2. What do the pod events say?
3. What do the deployment events say?

Uniquely to this operator: What is listed in the AIDeployment CRD?

## Getting help

Feel free to create an issue or reach out to us on [Prem's Discord](https://discord.com/invite/kpKk6vYVAn) etc.

## Scheduling Related

### Pod stuck in Pending

Often this is caused by a resource requirement. In particular GPU, but also CPU.

Even if you have exactly enough CPU and GPU to run your workload, Kubernetes won't be able to perform rolling updates.
You may have to manually scale down a deployment to stop the old pods and allow new ones to run. Alternatively
you can change Kubernetes update and Scheduling policies.

## GPU Related

Errors involving GPU drivers and runtime can be confusing. Sometimes the host's
kernel log can be helpful (`dmesg`). For instance the following message shows incompatibility between
the runtime and Kernel module/driver.

```
[  +0.056775] NVRM: API mismatch: the client has the version 545.23.08, but
              NVRM: this kernel module has the version 550.54.14.  Please
              NVRM: make sure that this kernel module and all NVIDIA driver
              NVRM: components have the same version.
```

### Pod won't start

1. Does the container have the correct version of CUDA?

Check the pod events for the following message or similar
`nvidia-container-cli.real: requirement error: invalid expression: unknown`

2. Does the container have a GPU specified?

If not this can manifest in commands (e.g. nvidia-smi) being absent.
