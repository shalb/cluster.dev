output "argocd_url" {
  value = "https://${local.argocd_domain}"
}

output "argocd_user" {
  value = "admin"
}

output "argocd_pass" {
  value = random_password.argocd_pass.result
}

output "keycloak_url" {
  value = "https://${local.keycloak_domain}"
}

output "keycloak_credentials" {
  value = data.kubernetes_secret.keycloak_credentials.data
}
