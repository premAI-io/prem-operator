provider "kubernetes" {
  config_context = "docker-desktop"
  config_path = "~/.kube/config"
}

resource "kubernetes_manifest" "simple-deployment" {
  manifest = {
    apiVersion = "premlabs.io/v1alpha1"
    kind       = "AIDeployment"

    metadata = {
      name = "deepspeed-mii"
      namespace = "prem-operator"
    }

    spec = {
      engine = {
        name = "deepspeed-mii"
      }
      models = [ { uri = "microsoft/phi-2" } ]
    }
  }
}