data "digitalocean_kubernetes_versions" "k8s" {
  version_prefix = "${var.k8s_version}."
}

resource "digitalocean_kubernetes_cluster" "k8s" {
  name    = var.cluster_name
  region  = var.region
  version = data.digitalocean_kubernetes_versions.k8s.latest_version
  vpc_uuid = data.terraform_remote_state.vpc.outputs.vpc_id

  node_pool {
    name       = "${var.cluster_name}-worker-pool"
    size       = var.node_type
    auto_scale = true
    min_nodes  = var.min_node_count
    max_nodes  = var.max_node_count
  }
}
