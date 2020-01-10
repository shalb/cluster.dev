# Generate private key to ssh to instance
resource "tls_private_key" "bastion_key" {
  algorithm   = "RSA"
}
# Generate file with key to pass to minikube module
#resource "local_file" "bastion_key" {
#    content     = tls_private_key.bastion_key.public_key_openssh
#    filename = "~/.ssh/id_rsa.pub"
#}

output "bastion_public_key_openssh" {
  value = tls_private_key.bastion_key.public_key_openssh
}

output "bastion_private_key_pem" {
  value = tls_private_key.bastion_key.private_key_pem
}