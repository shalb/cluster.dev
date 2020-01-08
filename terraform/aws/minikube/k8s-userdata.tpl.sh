# Install AWS utils
yum install -y wget unzip python3-pip.noarch
pip3 install awscli
export PATH=\$$PATH:/usr/local/bin

# copy kubeconfig to 
aws s3 cp /home/centos/kubeconfig s3://${cluster_name}/kubeconfig_${cluster_name}