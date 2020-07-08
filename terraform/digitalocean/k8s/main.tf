data "digitalocean_kubernetes_versions" "k8s" {
  version_prefix = "${var.k8s_version}."
}

resource "digitalocean_kubernetes_cluster" "k8s" {
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
