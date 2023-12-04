## Run Mistral on K3s with vLLM from saas-controller source code

### Overview
This guide will help you to run Mistral on K3s with vLLM from saas-controller source code.
Ubuntu 22.04.3 LTS was used as a host OS.

### Prerequisites
- K3s cluster
- Helm
- Git
- Go
- Docker
- make, curl, jq
- K9s(optional)
- Nvidia GPU Operator

### Steps
1. Install K3s cluster
```bash
./../scripts/install_k3s.sh
```
2. Install Helm
```bash
./.../scripts/install_helm.sh
```
3. Install Nvidia GPU Operator
```bash
./../scripts/install_gpu_operator_k3s.sh
```
4. Install tools: make, curl, jq
```bash
./../scripts/install_make_curl_jq.sh
```
5. Install Go
```bash
./../scripts/install_go.sh
```
6. Install Docker
```bash
./../scripts/install_docker.sh
```
7. Install K9s(optional)
```bash
./../scripts/install_k9s.sh
```
8. Clone saas-controller repository
```bash
git clone git@github.com:premAI-io/saas-controller.git
```
9. Deploy AIDeployment CRD
```bash
sudo make install
```
10. Build saas-controller Docker image
```bash
sudo make docker-build
```
11. Load Docker image to K3s cluster
```bash
sudo docker save -o ./controller controller:latest
sudo k3s ctr images import controller
```
12. Deploy saas-controller
```bash
sudo make deploy
```
13. Deploy vLLM
```bash
sudo kubectl apply -f ./../examples/vllm.yaml
```
14. Send request to vLLM
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
