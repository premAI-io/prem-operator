# Run Mistral on K3s with vLLM from prem-operator source code

## Overview

This guide will help you to run Mistral on K3s with vLLM from prem-operator source code.
Ubuntu 22.04.3 LTS was used as a host OS.

## Prerequisites

- K3s cluster
- Helm
- Git
- Go
- Docker
- make, curl, jq
- K9s(optional)
- Nvidia GPU Operator

## Steps

```bash
# 1. Install K3s cluster
./../scripts/install_k3s.sh

# 2. Install Helm
./.../scripts/install_helm.sh

# 3. Install Nvidia GPU Operator
./../scripts/install_gpu_operator_k3s.sh

# 4. Install tools: make, curl, jq
./../scripts/install_make_curl_jq.sh

# 5. Install Go
./../scripts/install_go.sh

# 6. Install Docker
./../scripts/install_docker.sh

# 7. Install K9s(optional)
./../scripts/install_k9s.sh

# 8. Clone prem-operator repository
git clone git@github.com:premAI-io/prem-operator.git

# 9. Deploy AIDeployment CRD
sudo make install

# 10. Build prem-operator Docker image
sudo make docker-build

# 11. Load Docker image to K3s cluster
sudo docker save -o ./controller controller:latest
sudo k3s ctr images import controller

# 12. Deploy prem-operator
sudo make deploy

# 13. Deploy vLLM
sudo kubectl apply -f ./../examples/vllm.yaml
```

### Send request to vLLM using curl and process the response with jq

```bash
curl -X 'POST' http://vllm.127.0.0.1.nip.io/v1/completions \
-H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "model": "mistralai/Mistral-7B-v0.1",
  "prompt": [
    "Nikola Tesla was special because"
  ],
  "max_tokens": 16,
  "temperature": 1,
  "top_p": 1,
  "n": 1,
  "stream": false,
  "logprobs": 0,
  "echo": false,
  "stop": [
    "string"
  ],
  "presence_penalty": 0,
  "frequency_penalty": 0,
  "best_of": 1,
  "user": "string",
  "top_k": -1,
  "ignore_eos": false,
  "use_beam_search": false,
  "stop_token_ids": [
    0
  ],
  "skip_special_tokens": true,
  "spaces_between_special_tokens": true
}' | jq -r '.choices[0].text'
```
