data "digitalocean_kubernetes_versions" "k8s" {
  version_prefix = "${var.k8s_version}."
}

resource "digitalocean_kubernetes_cluster" "k8s" {
  count   = var.enable_autoscaling ? 0 : 1
  name    = var.name
  region  = var.region
  version = data.digitalocean_kubernetes_versions.k8s.latest_version

  node_pool {
    name       = "${var.name}-worker-pool"
    size       = var.node_type
    node_count = var.node_count
  }
}

resource "digitalocean_kubernetes_cluster" "k8s_autoscaling" {
  count   = var.enable_autoscaling ? 1 : 0
  name    = var.name
  region  = var.region
  version = data.digitalocean_kubernetes_versions.k8s.latest_version

  node_pool {
    name       = "${var.name}-worker-pool"
    size       = var.node_type
    auto_scale = true
    min_nodes  = var.min_node_count
    max_nodes  = var.max_node_count
  }
}
