# Setup custom ingress provider

## Traefik

- Install traefik.

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik -n traefik --create-namespace
```

- Expose the traefik dashboard, if needed.

```bash
kubectl port-forward $(kubectl get pods --selector "app.kubernetes.io/name=traefik" --output=name) 9000:9000
```

- Use the annotations to configure the prem-operator ingress using traefik.

```yaml
apiVersion: premlabs.io/v1alpha1
kind: AIDeployment 
metadata:
  name: simple
spec:
  engine:
    name: "localai" 
  endpoint:
    - domain: "simple.127.0.0.1.nip.io"
      port: 8080 
  ingress:
    annotations:
      # use for adding annotations 
      traefik.ingress.kubernetes.io/router.entrypoints: web
  models:
    - uri: tinyllama-chat
```

## Nginx

- Install nginx ingress operator, [docs](https://kubernetes.github.io/ingress-nginx/deploy).

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx -n ingress-nginx --create-namespace
```

- Annotations to configure the prem-operator ingress using nginx.

```yaml
apiVersion: premlabs.io/v1alpha1
kind: AIDeployment 
metadata:
  name: simple
spec:
  engine:
    name: "localai" 
  endpoint:
    - domain: "simple.127.0.0.1.nip.io"
      port: 8080 
  ingress:
    annotations:
      # add annotations for nginx 
      nginx.ingress.kubernetes.io/auth-type: basic
      nginx.ingress.kubernetes.io/auth-secret: ingress-auth 
  models:
    - uri: tinyllama-chat
```
