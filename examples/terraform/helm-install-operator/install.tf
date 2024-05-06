provider "kubernetes" {
  config_context = "docker-desktop"
  config_path = "~/.kube/config"
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }

  # private registries can also be installed if needed.
  # registry {
  #   url = "oci://private.registry"
  #   username = "username"
  #   password = "password"
  # }
}

resource "kubernetes_namespace" "prem-operator" {
  metadata {
    name = "prem-operator"
  }
}

resource "helm_release" "prem-operator" {
  name        = "prem-op"
  namespace   = "prem-operator"
  repository  = "oci://registry-1.docker.io/premai"
  chart       = "prem-operator-chart"

  # depend on creation of namespace
  depends_on = [
    kubernetes_namespace.prem-operator
  ]

  #set {
  #  name  = "annotation-name"
  #  value = "annotation-value"
  #}
}
