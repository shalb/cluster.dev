# Generate private key to ssh to instance
resource "tls_private_key" "bastion_key" {
  algorithm   = "RSA"
}
# Generate file with key to pass to minikube module
#resource "local_file" "bastion_key" {
#    content     = tls_private_key.bastion_key.public_key_openssh
#    filename = "~/.ssh/id_rsa.pub"
#}