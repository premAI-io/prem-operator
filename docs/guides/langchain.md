# Building a langchain AI bot for conversation using Prem-Operator

## Requirements

- Kubernetes deployment with atleast 1 GPU based node, if you are using CPU based node consider using smaller models
- Prem Operator v0.1.0 or above

## Setup

1. Install Prem Operator
2. Install GPU Node if not already present.
3. Create a namespace for your deployments or you can just use default
4. Make sure that kubectl is installed on your machine, and KUBE_CONFIG for your cluster is setup as the default context.

## Getting the OpenAI Compatible AI Model

For this demonstration I will use a docker image with mosec serving library and torch for inference, you can consider using localai, vLLM or DeepSpeed-MII.
Or a custom image if you have one.

- Deploy the AI model with the Prem Operator

```bash
kubectl apply -f aideployment.yaml
```

```yaml
# aideployment.yaml
apiVersion: premlabs.io/v1alpha1
kind: AIDeployment
metadata:
    name: aideployment
spec:
    replicas: 1 # if you have more resources you can increase the replicas and the models will load balance the requests
    engine:
        name: generic 
    endpoint:
        # note: if you are using ingress for the deployment you would also
        # need to setup annotations checkout the ingress guide for it.
        - domain: "<DOMAIN-HERE>" # "langchain-model.com"
          port: 8000
    deployment:
        template:
            spec:
                containers:
                    - name: "ai-model-7b"
                      image: modelzai/llm-llama-7b:latest # GPU requirements: A100(40GB)
                      # alternatively use, modelzai/llm-bloomz-560m:latest (even works on CPU)
        accelerator:
            interface: "CUDA"
            minVersion:
                major: 7
        resources:
            limits:
                cpu: "1"
                memory: "16Gi" # a decent amount of RAM is required for loading the model as well
```

- Port forward the deployment service. If you don't have a proper ingress setup for your cluster.

```bash
kubectl port-forward service/aideployment 8000:8000
```

- Now locally, install the required libraries.

```bash
pip install langchain openai
```

- Create a python script and import the openai library. Setup the API against your newly deployed model.

```python
import os
import openai

# note: we port-forwarded the service to 8000
openai.api_base="http://localhost:8000"
# if you have ingress setup then use your domain name
# you can also modify the port to use http(s) port itself
# openai.api_base="https://<DOMAIN-NAME>.tld:8000" 
openai.api_key = "any"
```

- Now you can setup the rest of the langchain.

```python
import os
import openai

# note: we port-forwarded the service to 8000
openai.api_base="http://localhost:8000"
openai.api_key = "any"

from langchain.llms import OpenAI
from langchain.chains import ConversationChain
from langchain.memory import ConversationBufferMemory

# Create a Conversation Chain
llm = OpenAI(temperature=0)
memory = ConversationBufferMemory()
conversation = ConversationChain(llm=llm, memory=memory)

# Interact with the Chain
while True:
    query = input("User: ")
    result = conversation.run(query)
    print(f"Assistant: {result}")
```

- Run the project to the demo in action.

```bash
python demo.py
```
