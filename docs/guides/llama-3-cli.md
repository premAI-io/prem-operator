## Llama 3 TUI and CLI

Chat with Llama 3 in [Elia TUI](https://github.com/darrenburns/elia) or use it on the command line with the [llm](https://github.com/simonw/llm) CLI.

### Quick Start on GPU

1. Install the NVIDIA Operator 
2. Install the Prem Operator
    ```
    $ helm install latest oci://registry-1.docker.io/premai/prem-operator-chart
    ```

Now you have a choice of different models.

### Llama 3 70b

The Llama 3 70b model we are using in the example YAML requires between 40GB and 50GB of GPU memory.

3. Deploy the 70b model and associated tools
    ```
    $ kubectl apply -f examples/llama3-70b-gguf.yaml
    ```

### Llama 3 8b

The Llama 3 8b model we are using in the example requires about 8GB of GPU memory.

3. Deploy the 8b model and associated tools
    ```
    $ kubectl apply -f examples/llama3-8b-gguf.yaml
    ```

### Generate a script with llm

```
$ kubectl exec -it deployments/llama-3-cli -- llm -s "You are a code refactoring and generation tool. You only output valid code. Do not include triple quotes or markdown that wraps the code" "Write a script tells me if it is monday" | tee ismonday.py
$ python ismonday.py
```

Note that you can pipe files into `kubectl` and they are appended to the prompt in the arguments. However it is easy to exceed Llama 2's context window.

### Interactive chat with Elia

```
$ kubectl exec -it deployments/llama-3-tui elia
```

Select GPT 4 as the model.

