# Managed Kubernetes Cluster

## Basics

Install [k8s](https://kubernetes.io/docs/tasks/tools/) tools, i.e., kubectl.

### AWS CLI

- Install aws cli tools. `python -m pip install awscli`
- Configure it `aws configure`. You can also use shared credentials file, `~/.aws/credentials` or `%USERPROFILE%\.aws\credentials`.

    ```ini
    [default]
    aws_access_key_id=MYACCESSKEY
    aws_secret_access_key=MYSECRETKEY
    # Optional, to define default region for this profile.
    region=us-west-1

    [testing]
    aws_access_key_id=MYACCESSKEY
    aws_secret_access_key=MYSECRETKEY
    ```

- Create the cluster. `aws eks create-cluster --name hello-cluster --region us-east-1`
- To get access via kube config use `aws eks update-kubeconfig --region us-east-1 --name hello-cluster`.

Note: For GPU nodes you have to create a nodegroup specifically for gpus with say p3 or p5 instance type.

#### More information on setting up production clusters in AWS

- Setting up VPC: [here](https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html)
- And IAM for the cluster: [here](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-iam.html)

### GCP CLI

- Install [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/docs/) tools, i.e., gcloud and kubectl.
- Config [gcloud](https://cloud.google.com/sdk/gcloud). `gcloud config set project PROJECT_ID`
- Create the cluster. ` gcloud container clusters create-auto hello-cluster --location=us-central1 `
- Setup kube config for cluster to be able to use tools like kubectl, with `gcloud container clusters get-credentials hello-cluster`.
- Now you can follow the steps from the getting started page.

Note: For GPU nodes you have to create a nodegroup specifically for gpus with say p3 or p5 instance type.

#### More information on setting up production clusters in GCP

- IAM: [Access Control](https://cloud.google.com/vpc-service-controls/docs/access-control)
- VPC: [here](https://cloud.google.com/vpc/docs/using-vpc)
- Best Practices for VPC: [here](https://cloud.google.com/vpc-service-controls/docs/enable)

### Using Terraform

Requires the cluster tools for the platform installed, in this case either AWS or GCP.
We are going to be using GCP in the following code.

Definition for the cluster in terraform.

```tf
resource "google_container_cluster" "gpu_cluster" {
  name           = var.cluster_name
  location       = var.region
  node_locations = [var.zone]

  deletion_protection = false
  initial_node_count  = 1

  timeouts {
    create = "20m"
    update = "30m"
  }

  lifecycle {
    ignore_changes = [master_auth, network]
  }
}
```

GPU node pool for the cluster(You can add more gpus here), this adds the non master gpu nodes.

```tf
resource "google_container_node_pool" "gpu_pool" {
  name       = "gpu-pool"
  location   = var.region
  cluster    = google_container_cluster.gpu_cluster.name
  node_count = 1


  management {
    auto_repair  = "true"
    auto_upgrade = "true"
  }

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/trace.append",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
    ]


    labels = {
      env     = "sandbox"
      project = var.project_id

    }

    guest_accelerator {
      type  = var.gpu_type
      count = 1
      gpu_driver_installation_config {
        gpu_driver_version = var.gpu_driver_version
      }

    }

    image_type   = "cos_containerd"
    machine_type = "g2-standard-24"
    tags         = ["gke-node", "sandbox", "${var.project_id}"]

    disk_size_gb = "50"
    disk_type    = "pd-balanced"

    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }


  }

  timeouts {
    create = "20m"
    update = "30m"
  }
}
```

Now you can access the cluster using kubectl by running `gcloud container clusters get-credentials $CLUSTER_NAME --region $REGION`
