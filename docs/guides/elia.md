## Elia TUI

Deploy the operator and chat with Phi-2 in the [Elia TUI](https://github.com/darrenburns/elia).

![elia](https://github.com/richiejp/prem-operator/assets/988098/e54eb9b8-1375-4b35-ba3e-24ba5d58360a)

### Quick Start on GPU

1. Install the NVIDIA Operator (optional with Phi-2 see below)
2. Install the Prem Operator
    ```
    $ helm install latest oci://registry-1.docker.io/premai/prem-operator-chart
    ```

3. Deploy the Phi-2 + Elia example and exec Elia in a terminal session
    ```
    $ kubectl apply -f examples/elia-tui.yaml
    $ kubectl exec deployments/elia elia
    ```

### Quick Start on CPU

1. Install the Prem Operator
    ```
    $ helm install latest oci://registry-1.docker.io/premai/prem-operator-chart
    ```

2. Deploy the Phi-2 + Elia CPU example and exec Elia in a terminal session
    ```
    $ kubectl apply -f examples/elia-tui-cpu.yaml
    $ kubectl exec deployments/elia elia
    ```

