# Install AWS utils
apt-get update
apt-get install -y wget unzip python3-pip
pip3 install awscli
export PATH=\$$PATH:/usr/local/bin

#copy kubeconfig to s3
aws s3 cp /home/ubuntu/kubeconfig s3://${cluster_name}/kubeconfig_${cluster_name}

copy private ssh key to s3
cat << EOF > /home/ubuntu/.ssh/id_rsa
EOF
chmod 600 /home/ubuntu/.ssh/id_rsa
aws s3 cp /home/ubuntu/.ssh/id_rsa  s3://${cluster_name}/id_rsa_${cluster_name}.pem
