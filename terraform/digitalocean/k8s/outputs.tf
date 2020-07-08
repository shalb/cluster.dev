output "cluster_status" {
  value = digitalocean_kubernetes_cluster.k8s.status
}

output "cluster_endpoint" {
  value = digitalocean_kubernetes_cluster.k8s.endpoint
}
